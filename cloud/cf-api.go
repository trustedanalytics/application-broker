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
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cloudfoundry-community/types-cf"
	"github.com/nu7hatch/gouuid"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"github.com/twmb/algoimpl/go/graph"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/url"
	"strings"
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

// Returns a list of services and apps which would be provisioned in normal run
func (c *CfAPI) DryRun(sourceAppGUID string) ([]types.Component, error) {
	sourceAppSummary, err := c.getAppSummary(sourceAppGUID)
	if err != nil {
		return nil, err
	}

	g := graph.New(graph.Directed)
	root := g.MakeNode()
	*root.Value = types.Component{
		GUID:         sourceAppGUID,
		Name:         sourceAppSummary.Name,
		Type:         types.ComponentApp,
		DependencyOf: []string{},
		Clone:        true,
	}

	_ = c.addDependenciesToGraph(g, root, sourceAppGUID)
	// Calculations
	sorted := g.TopologicalSort()
	log.Infof("Topological Order:\n")
	ret := make([]types.Component, len(sorted))
	for i, node := range sorted {
		text := ""
		for _, n := range g.Neighbors(node) {
			text += fmt.Sprint(*n.Value) + ","
		}
		log.Infof("%v [%v]", *node.Value, text)
		ret[len(sorted)-1-i] = (*node.Value).(types.Component)
	}
	ret = c.removeDuplicates(ret)

	if c.graphHasCycles(g) {
		log.Errorf("Graph has cycles and stack cannot be copied")
		return nil, misc.InternalServerError{Context: ""}
	} else {
		log.Infof("Graph has no cycles")
	}
	return ret, nil
}

func (c *CfAPI) removeDuplicates(components []types.Component) []types.Component {
	m := make(map[string]int)
	var ret []types.Component
	for _, n := range components {
		j, ok := m[n.GUID]
		if !ok {
			ret = append(ret, n)
			m[n.GUID] = len(ret) - 1
			continue
		} else {
			log.Infof("Duplicated %v on position %v", n.Name, j)
			ret[j].DependencyOf = append(ret[j].DependencyOf, n.DependencyOf...)
			log.Infof("Merged dependencies %v", ret[j].DependencyOf)
		}
	}
	return ret
}

func (c *CfAPI) graphHasCycles(g *graph.Graph) bool {
	components := g.StronglyConnectedComponents()
	for _, comp := range components {
		if len(comp) > 1 {
			return true
		}
	}
	return false
}

func (c *CfAPI) addDependenciesToGraph(g *graph.Graph, parent graph.Node, sourceAppGUID string) error {
	log.Infof("addDependenciesToGraph for parent %v", *parent.Value)
	sourceAppSummary, err := c.getAppSummary(sourceAppGUID)
	if err != nil {
		return err
	}
	for _, svc := range sourceAppSummary.Services {
		node := g.MakeNode()
		if c.isNormalService(svc) {
			*node.Value = types.Component{
				GUID:         svc.GUID,
				Name:         svc.Name,
				Type:         types.ComponentService,
				DependencyOf: []string{(*parent.Value).(types.Component).GUID},
				Clone:        true,
			}
			g.MakeEdgeWeight(parent, node, 1)
		} else {
			*node.Value = types.Component{
				GUID:         svc.GUID,
				Name:         svc.Name,
				Type:         types.ComponentUPS,
				DependencyOf: []string{(*parent.Value).(types.Component).GUID},
				Clone:        true,
			}
			g.MakeEdgeWeight(parent, node, 1)
			// Retrieve UPS
			response, err := c.getUserProvidedService(svc.GUID)
			if err != nil {
				return err
			}
			if val, ok := response.Entity.Credentials["url"]; ok {
				if urlStr, ok := val.(string); ok {
					appID, appName, err := c.getAppIdAndNameFromSpaceByUrl(sourceAppSummary.SpaceGUID, urlStr)
					if err != nil {
						return err
					}
					if len(appID) > 0 {
						log.Infof("Application %v is bound using %v", appID, svc.Name)
						node2 := g.MakeNode()
						*node2.Value = types.Component{
							GUID:         appID,
							Name:         appName,
							Type:         types.ComponentApp,
							DependencyOf: []string{(*node.Value).(types.Component).GUID},
							Clone:        true,
						}
						g.MakeEdgeWeight(node, node2, 1)
						_ = c.addDependenciesToGraph(g, node2, appID)
					}
				}
			}
		}
	}
	return nil
}

// Provision instantiates service of given type
func (c *CfAPI) Provision(sourceAppGUID string, r *cf.ServiceCreationRequest) (*types.ServiceCreationResponse, error) {
	order, err := c.DryRun(sourceAppGUID) // TEST EXECUTION
	log.Infof("Dry run: [%v]", order)

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

	asyncErr := make(chan error)
	go c.copyBits(sourceAppGUID, destApp.Meta.GUID, asyncErr)

	route, err := c.createRoute(&types.CfCreateRouteRequest{requestedName, domainGUID, r.SpaceGUID})
	if err != nil {
		return nil, err
	}

	if err := c.associateRoute(destApp.Meta.GUID, route.Meta.GUID); err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(sourceAppSummary.Services))
	results := make(chan error, len(sourceAppSummary.Services))
	// Create dependent services and bind them
	for _, svcToCopy := range sourceAppSummary.Services {
		go c.createDependencies(destApp, svcToCopy, &wg, results)
	}
	wg.Wait()
	if err := misc.FirstNonEmpty(results, len(sourceAppSummary.Services)); err != nil {
		return nil, err
	}

	//Waiting for copy_bits finish
	err = <-asyncErr
	if err != nil {
		return nil, err
	}

	if err := c.startApp(destApp); err != nil {
		return nil, err
	}

	log.Infof("Service instance [%v] created", requestedName)
	destApp.Meta.URL = fmt.Sprintf("%s.%s", route.Entity.Host, domainName)
	toReturn := types.ServiceCreationResponse{
		App: *destApp,
		ServiceCreationResponse: cf.ServiceCreationResponse{DashboardURL: ""},
	}
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

func (c *CfAPI) createDependencies(destApp *types.CfAppResource, svc types.CfAppSummaryService, wg *sync.WaitGroup, errors chan error) {
	defer wg.Done()
	serviceNameGuid, _ := uuid.NewV4()
	serviceName := svc.Name + "-" + serviceNameGuid.String()[:8]
	log.Debugf("Create dependent service: service=[%v] ([%v], [%v])", serviceName, svc.Plan.Service.Label, svc.Plan.Name)

	var spawnedServiceInstanceGUID string
	if c.isNormalService(svc) {
		// svc is a normal service
		// Create service
		svcInstanceReq := types.NewCfServiceInstanceRequest(serviceName, destApp.Entity.SpaceGUID, svc.Plan)
		response, err := c.createServiceInstance(svcInstanceReq)
		if err != nil {
			errors <- err
			return
		}
		spawnedServiceInstanceGUID = response.Meta.GUID
		log.Debugf("Dependent service created: Service Instance GUID=[%v]", spawnedServiceInstanceGUID)
	} else {
		// svc is a user provided service
		// Retrieve UPS
		response, err := c.getUserProvidedService(svc.GUID)
		if err != nil {
			errors <- err
			return
		}
		log.Infof("Dependent user provided service retrieved: %+v", response)
		// Create UPS
		response.Entity.Name = serviceName
		response.Entity.SpaceGUID = destApp.Entity.SpaceGUID

		if val, ok := response.Entity.Credentials["url"]; ok {
			if urlStr, ok := val.(string); ok {
				appID, appName, err := c.getAppIdAndNameFromSpaceByUrl(destApp.Entity.SpaceGUID, urlStr)
				if err != nil {
					errors <- err
					return
				}
				log.Infof("Application %v (%v) is bound using %v", appID, appName, svc.Name)
			}
		}

		_ = c.applyAdditionalReplacementsInUPSCredentials(response)

		response, err = c.createUserProvidedServiceInstance(&response.Entity)
		if err != nil {
			errors <- err
			return
		}
		spawnedServiceInstanceGUID = response.Meta.GUID
		log.Debugf("Dependent user provided service created. Service Instance GUID=[%v]", spawnedServiceInstanceGUID)

	}
	// Bind created service
	svcBindingReq := types.NewCfServiceBindingRequest(destApp.Meta.GUID, spawnedServiceInstanceGUID)
	svcBindingResp, err := c.createServiceBinding(svcBindingReq)
	if err != nil {
		errors <- err
		return
	}
	log.Debugf("Dependent service binding created: Service Binding GUID=[%v]", svcBindingResp.Meta.GUID)

	errors <- nil
	return
}

func (c *CfAPI) isNormalService(svc types.CfAppSummaryService) bool {
	// Normal services require plan.
	// User provided services does not support Plans so this field is empty then.
	return len(svc.Plan.Service.Label) > 0
}

func (c *CfAPI) doesUrlMatchApplication(appUrlStr, appID string) (bool, error) {
	appURL, err := url.Parse(appUrlStr)
	if err != nil {
		return false, err
	}
	appSummary, err := c.getAppSummary(appID)
	log.Infof("App summary retrieved is [%+v]", appSummary)
	if err != nil {
		return false, err
	}
	for i := range appSummary.Routes {
		route := appSummary.Routes[i]
		if appURL.Host == fmt.Sprintf("%v.%v", route.Host, route.Domain.Name) {
			return true, nil
		}
	}
	return false, nil
}

func (c *CfAPI) applyAdditionalReplacementsInUPSCredentials(response *types.CfUserProvidedServiceResource) error {
	credentials, err := json.Marshal(response.Entity.Credentials)
	if err != nil {
		return err
	}
	credentialsStr := string(credentials)
	credentialsStr = misc.ReplaceWithRandom(credentialsStr)
	json.Unmarshal([]byte(credentialsStr), &response.Entity.Credentials)
	return nil
}

func (c *CfAPI) getAppIdAndNameFromSpaceByUrl(spaceGUID, urlStr string) (string, string, error) {
	appURL, err := url.Parse(urlStr)
	if err != nil {
		log.Infof("[%v] is not a correct URL. Parsing failed.", urlStr)
		return "", "", err
	}
	log.Infof("URL Host %v", appURL.Host)
	routes, err := c.getSpaceRoutesForHostname(spaceGUID, strings.Split(appURL.Host, ".")[0])
	if err != nil {
		return "", "", err
	}
	if routes.Count > 0 {
		log.Infof("%v route(s) retrieved for host %v", routes.Count, appURL.Host)
		routeGUID := routes.Resources[0].Meta.GUID
		apps, err := c.getAppsFromRoute(routeGUID)
		if err != nil {
			return "", "", err
		}
		if apps.Count > 0 {
			app := apps.Resources[0]
			log.Infof("APP [%+v]", app)
			isSearched, err := c.doesUrlMatchApplication(urlStr, app.Meta.GUID)
			if err != nil {
				return "", "", err
			}
			if isSearched {
				log.Infof("Found app match url in user provided service")
				return app.Meta.GUID, app.Entity.Name, nil
			} else {
				log.Infof("url of found app does not match url in user provided service")
			}
		} else {
			log.Infof("No apps bound to route: [%v]", routeGUID)
		}
	} else {
		log.Infof("No routes found for host: %v", appURL.Host)
	}
	return "", "", nil
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
