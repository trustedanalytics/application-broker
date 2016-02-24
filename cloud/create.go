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
	"github.com/juju/errors"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"strings"
	"sync"
)

func (cl *CloudAPI) createService(spaceGUID string, comp types.Component, suffix string,
	wg *sync.WaitGroup, results chan types.ComponentClone, errorsCh chan error) {

	defer wg.Done()

	if len(comp.DependencyOf) == 0 {
		errorsCh <- errors.New("Service not attached to any application")
	}
	parentApp, err := cl.cf.GetAppSummary(comp.DependencyOf[0])
	if err != nil {
		errorsCh <- err
		return
	}

	var svc types.CfAppSummaryService
	for _, s := range parentApp.Services {
		if s.GUID == comp.GUID {
			svc = s
		}
	}

	serviceName := svc.Name + "-" + suffix
	log.Debugf("Create dependent service: service=[%v] ([%v], [%v])", serviceName, svc.Plan.Service.Label, svc.Plan.Name)

	// Create service
	svcInstanceReq := types.NewCfServiceInstanceRequest(serviceName, spaceGUID, svc.Plan)
	response, err := cl.cf.CreateServiceInstance(svcInstanceReq)
	if err != nil {
		errorsCh <- err
		return
	}
	spawnedServiceInstanceGUID := response.Meta.GUID
	log.Debugf("Dependent service created: Service Instance GUID=[%v]", spawnedServiceInstanceGUID)

	results <- types.ComponentClone{
		Component: comp,
		CloneGUID: spawnedServiceInstanceGUID,
	}
	errorsCh <- nil
	return
}

func (cl *CloudAPI) createApplication(sourceAppGUID, spaceGUID string, parameters map[string]string) (*types.CfAppResource, error) {
	// Gather reference app summary to be used later for creating new instance
	sourceAppSummary, err := cl.cf.GetAppSummary(sourceAppGUID)
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
	requestedName := parameters["name"]
	delete(parameters, "name")

	if len(sourceAppSummary.Routes) == 0 {
		return nil, misc.InternalServerError{Context: "Reference app has no route associated"}
	}

	//Newly spawned app instance shall have almost identical config as reference app
	destApp := types.NewCfAppResource(*sourceAppSummary, requestedName, spaceGUID)
	if destApp.Entity.Envs == nil && len(parameters) > 0 {
		destApp.Entity.Envs = map[string]interface{}{}
	}
	for k, v := range parameters {
		log.Debugf("Setting additional env: %v:%v", k, v)
		if _, ok := destApp.Entity.Envs[k]; ok {
			log.Warnf("Env %v already exists (overriding)", k)
		}
		destApp.Entity.Envs[k] = v
	}
	destApp, err = cl.cf.CreateApp(destApp.Entity)
	if err != nil {
		return nil, err
	}

	domainGUID := sourceAppSummary.Routes[0].Domain.GUID
	domainName := sourceAppSummary.Routes[0].Domain.Name

	route, err := cl.cf.CreateRoute(&types.CfCreateRouteRequest{requestedName, domainGUID, spaceGUID})
	if err != nil {
		return nil, err
	}

	if err := cl.cf.AssociateRoute(destApp.Meta.GUID, route.Meta.GUID); err != nil {
		return nil, err
	}

	destApp.Meta.URL = fmt.Sprintf("%s.%s", route.Entity.Host, domainName)
	return destApp, nil
}

func (cl *CloudAPI) createUserProvidedService(spaceGUID string, comp types.Component, suffix, url string,
	wg *sync.WaitGroup, results chan types.ComponentClone, errorsCh chan error) {

	defer wg.Done()
	serviceName := comp.Name + "-" + suffix
	log.Debugf("Create dependent user provided service: service=[%v])", serviceName)

	// Retrieve UPS
	response, err := cl.cf.GetUserProvidedService(comp.GUID)
	if err != nil {
		errorsCh <- err
		return
	}
	log.Infof("Dependent user provided service retrieved: %+v", response)

	// Create UPS
	response.Entity.Name = serviceName
	response.Entity.SpaceGUID = spaceGUID

	// Replace url to match clone application route
	if len(url) > 0 {
		response.Entity.Credentials["url"] = fmt.Sprintf("http://%v", url)
		response.Entity.Name = fmt.Sprintf("%v-ups", strings.Split(url, ".")[0])
	}
	// Generate random values where needed
	_ = cl.applyAdditionalReplacementsInUPSCredentials(response)

	response, err = cl.cf.CreateUserProvidedServiceInstance(&response.Entity)
	if err != nil {
		errorsCh <- err
		return
	}
	spawnedServiceInstanceGUID := response.Meta.GUID
	log.Debugf("Dependent user provided service created. Service Instance GUID=[%v]", spawnedServiceInstanceGUID)

	results <- types.ComponentClone{
		Component: comp,
		CloneGUID: spawnedServiceInstanceGUID,
	}
	errorsCh <- nil
	return
}
