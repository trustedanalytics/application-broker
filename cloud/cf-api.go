/**
 * Copyright (c) 2015 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cloud

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cloudfoundry-community/types-cf"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"sync"
)

// CfAPI is the implementation of API interface. It is point of access to CF CloudController API
type CfAPI struct {
	BaseAddress string
	*http.Client
}

// NewCfAPI constructs and initializes access to CF by loading necessary credentials from ENVs
func NewCfAPI() *CfAPI {
	envs := cfenv.CurrentEnv()
	tokenConfig := &clientcredentials.Config{
		ClientID:     envs["CLIENT_ID"],
		ClientSecret: envs["CLIENT_SECRET"],
		Scopes:       []string{},
		TokenURL:     envs["TOKEN_URL"],
	}
	toReturn := new(CfAPI)
	toReturn.BaseAddress = envs["CF_API"]
	toReturn.Client = tokenConfig.Client(oauth2.NoContext)
	return toReturn
}

// Provision instantiates service of given type
func (c *CfAPI) Provision(sourceAppGUID string, r *cf.ServiceCreationRequest) (*types.ServiceCreationResponse, error) {
	// Gather reference app summary to be used later for creating new instance
	sourceAppSummary, err := c.getAppSummary(sourceAppGUID)
	if err != nil {
		switch err {
		case misc.EntityNotFoundError{}:
			log.Errorf("Application %v not found", sourceAppGUID)
			return nil, misc.InternalServerError{Context: err.Error()}
		default:
			log.Errorf("Failed to get application summary: %v", err.Error())
			return nil, err
		}
	}
	requestedName := r.Parameters["name"]
	delete(r.Parameters, "name")

	if len(sourceAppSummary.Routes) == 0 {
		return nil, misc.InternalServerError{Context: "Reference app has no route associated"}
	}
	domainGUID := sourceAppSummary.Routes[0].Domain.GUID
	domainName := sourceAppSummary.Routes[0].Domain.Name

	//Newly spawned app instance shall have almost identical config as reference app
	destApp := types.NewCfAppResource(*sourceAppSummary, requestedName, r.SpaceGUID)
	if destApp.Entity.Envs == nil && len(r.Parameters) > 0 {
		destApp.Entity.Envs = map[string]interface{}{}
	}
	for k, v := range r.Parameters {
		log.Debugf("Setting additional env: %v:%v", k, v)
		if _, ok := destApp.Entity.Envs[k]; ok {
			log.Warnf("Env %v already exists (overriding)", k)
		}
		destApp.Entity.Envs[k] = v
	}
	destApp, err = c.createApp(destApp.Entity)
	if err != nil {
		return nil, err
	}

	toReturn := types.ServiceCreationResponse{
		App: destApp,
		ServiceCreationResponse: cf.ServiceCreationResponse{DashboardURL: ""},
	}

	asyncErr := make(chan error)
	go c.copyBits(sourceAppGUID, destApp.Meta.GUID, asyncErr)

	route, err := c.createRoute(&types.CfCreateRouteRequest{requestedName, domainGUID, r.SpaceGUID})
	if err != nil {
		return &toReturn, err
	}

	if err := c.associateRoute(destApp.Meta.GUID, route.Meta.GUID); err != nil {
		return &toReturn, err
	}

	destApp.Meta.URL = fmt.Sprintf("%s.%s", route.Entity.Host, domainName)

	wg := sync.WaitGroup{}
	wg.Add(len(sourceAppSummary.Services))
	results := make(chan error, len(sourceAppSummary.Services))
	// Create dependent services and bind them
	suffix := r.InstanceID[:8]
	for _, svcToCopy := range sourceAppSummary.Services {
		go c.createDependencies(destApp, svcToCopy, suffix, &wg, results)
	}
	wg.Wait()
	if err := misc.FirstNonEmpty(results, len(sourceAppSummary.Services)); err != nil {
		return &toReturn, err
	}

	//Waiting for copy_bits finish
	err = <-asyncErr
	if err != nil {
		return &toReturn, err
	}

	if err := c.startApp(destApp); err != nil {
		return &toReturn, err
	}

	log.Infof("Service instance [%v] created", requestedName)
	return &toReturn, nil
}

// Deprovision remove instance of given application (that stands behind service instance though)
func (c *CfAPI) Deprovision(appGUID string) error {
	summary, err := c.getAppSummary(appGUID)
	if err != nil {
		switch err {
		case misc.EntityNotFoundError{}:
			log.Warnf("Application %v not found. Aborting deprovision silently...", appGUID)
			return nil
		default:
			log.Errorf("Failed to get application summary: %v", err.Error())
			return err
		}
	}

	results := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go c.deleteBoundServices(appGUID, results, &wg)
	go c.deleteRoutes(appGUID, summary.Routes, results, &wg)

	wg.Wait()
	if err := misc.FirstNonEmpty(results, 2); err != nil {
		return err
	}

	if err := c.deleteApp(appGUID); err != nil {
		return err
	}

	return nil
}

// UpdateBroker registers or updates catalog in CF
func (c *CfAPI) UpdateBroker(brokerName string, brokerURL string, username string, password string) error {
	brokers, err := c.getBrokers(brokerName)
	if err != nil {
		return err
	}

	if brokers.TotalResults == 0 {
		return c.registerBroker(brokerName, brokerURL, username, password)
	}
	return c.updateBroker(brokers.Resources[0].Meta.GUID, brokerURL, username, password)
}

func (c *CfAPI) CheckIfServiceExists(serviceName string) error {
	myData := types.GetVcapApplication()
	broker, err := c.getBrokers(myData.Name)
	duplicate, err := c.getServiceOfName(serviceName)
	if err != nil {
		return err
	}
	if duplicate != nil {
		if broker.TotalResults == 0 || broker.Resources[0].Meta.GUID != duplicate.Entity.BrokerGUID {
			return misc.InternalServerError{Context: "Service name already registered in different CF broker!"}
		} else if broker.TotalResults > 0 && broker.Resources[0].Meta.GUID == duplicate.Entity.BrokerGUID {
			log.Infof("Service name was registered in CF for THIS broker but was missing in internal DB, purging...", serviceName)
			return c.purgeService(duplicate.Meta.GUID, duplicate.Entity.Name, duplicate.Entity.PlansURL)
		}
	}
	return nil
}

func (c *CfAPI) createDependencies(destApp *types.CfAppResource, svc types.CfAppSummaryService, suffix string, wg *sync.WaitGroup, errors chan error) {
	defer wg.Done()
	serviceName := svc.Name + "-" + suffix
	log.Debugf("Create dependent service: service=[%v] ([%v], [%v])", serviceName, svc.Plan.Service.Label, svc.Plan.Name)

	// Create service
	svcInstanceReq := types.NewCfServiceInstanceRequest(serviceName, destApp.Entity.SpaceGUID, svc.Plan)
	response, err := c.createServiceInstance(svcInstanceReq)
	if err != nil {
		errors <- err
		return
	}
	log.Debugf("Dependent service created: Service Instance GUID=[%v]", response.Meta.GUID)
	// Bind created service
	svcBindingReq := types.NewCfServiceBindingRequest(destApp.Meta.GUID, response.Meta.GUID)
	svcBindingResp, err := c.createServiceBinding(svcBindingReq)
	if err != nil {
		errors <- err
		return
	}
	log.Debugf("Dependent service binding created: Service Binding GUID=[%v]", svcBindingResp.Meta.GUID)
	errors <- nil
	return
}

func (c *CfAPI) deleteBoundServices(appGUID string, result chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()

	bindings, err := c.getAppBindigs(appGUID)
	if err != nil {
		result <- err
		return
	}
	var results = make(chan error, len(bindings.Resources))
	wg := sync.WaitGroup{}
	wg.Add(len(bindings.Resources))

	for _, loopBinding := range bindings.Resources {
		go func(binding types.CfBindingResource) {
			defer wg.Done()
			if err := c.deleteBinding(binding); err != nil {
				results <- err
				return
			}
			if err := c.deleteServiceInstance(binding.Entity.ServiceInstanceGUID); err != nil {
				results <- err
				return
			}
			results <- nil
		}(loopBinding)
	}
	wg.Wait()
	result <- misc.FirstNonEmpty(results, len(bindings.Resources))
}

func (c *CfAPI) deleteRoutes(appGUID string, routes []types.CfAppSummaryRoute, result chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()

	wg := sync.WaitGroup{}
	wg.Add(len(routes))
	results := make(chan error, len(routes))

	for _, loopRoute := range routes {
		go func(route types.CfAppSummaryRoute) {
			defer wg.Done()
			if err := c.unassociateRoute(appGUID, route.GUID); err != nil {
				results <- err
				return
			}
			if err := c.deleteRoute(route.GUID); err != nil {
				results <- err
				return
			}
			results <- nil
		}(loopRoute)
	}
	wg.Wait()
	result <- misc.FirstNonEmpty(results, len(routes))
}
