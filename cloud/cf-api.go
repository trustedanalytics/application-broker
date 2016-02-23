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
	"github.com/juju/errors"
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
		return nil, misc.InternalServerError{Context: "Graph has cycles and stack cannot be copied"}
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
	order, _ := c.DryRun(sourceAppGUID)
	log.Infof("Dry run: [%v]", order)
	log.Infof("%v components to spawn:", len(order))

	componentsToSpawn := make(map[types.ComponentType][]types.Component)
	componentsToSpawn[types.ComponentApp] = []types.Component{}
	componentsToSpawn[types.ComponentUPS] = []types.Component{}
	componentsToSpawn[types.ComponentService] = []types.Component{}
	for _, comp := range order {
		if _, ok := componentsToSpawn[comp.Type]; ok {
			componentsToSpawn[comp.Type] = append(componentsToSpawn[comp.Type], comp)
		} else {
			componentsToSpawn[comp.Type] = []types.Component{comp}
		}
	}
	log.Infof("%+v", componentsToSpawn)

	commonUUID, _ := uuid.NewV4()
	suffix := commonUUID.String()[:8]

	destAppsResources := make(map[string]*types.CfAppResource)

	log.Infof("Creating main application")
	destApp, err := c.createApplication(sourceAppGUID, r.SpaceGUID, r.Parameters)
	if err != nil {
		return nil, err
	}
	destAppsResources[sourceAppGUID] = destApp
	log.Infof("Creating dependent applications")
	for _, app := range componentsToSpawn[types.ComponentApp] {
		if _, ok := destAppsResources[app.GUID]; !ok {
			name := fmt.Sprintf("%v-%v", app.Name, suffix)
			params := map[string]string{"name": name}
			appRes, err := c.createApplication(app.GUID, r.SpaceGUID, params)
			if err != nil {
				return nil, err
			}
			destAppsResources[app.GUID] = appRes
		}
	}

	log.Infof("Copying applications data")
	copyBitsAsyncErrors := make(chan error, len(destAppsResources))
	for _, appRes := range destAppsResources {
		go c.copyBits(sourceAppGUID, appRes.Meta.GUID, copyBitsAsyncErrors)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(componentsToSpawn[types.ComponentService]))
	errors := make(chan error, len(componentsToSpawn[types.ComponentService]))
	results := make(chan types.ComponentClone, len(componentsToSpawn[types.ComponentService]))
	// Create dependent services
	log.Infof("Creating dependent services")
	required_bindings := 0
	for _, comp := range componentsToSpawn[types.ComponentService] {
		required_bindings += len(comp.DependencyOf)
		go c.createService(destApp.Entity.SpaceGUID, comp, suffix, &wg, results, errors)
	}
	wg.Wait()
	close(errors)
	close(results)
	if err := misc.FirstNonEmpty(errors, len(componentsToSpawn[types.ComponentService])); err != nil {
		return nil, err
	}
	log.Infof("Required bindings: %v", required_bindings)
	wg.Add(required_bindings)
	errorsBind := make(chan error, required_bindings)
	// Bind services
	log.Infof("Binding dependent services")
	for clone := range results {
		for _, dependent := range clone.Component.DependencyOf {
			go c.bindService(destAppsResources[dependent].Meta.GUID, clone.CloneGUID, &wg, errorsBind)
		}
	}
	wg.Wait()
	close(errorsBind)
	if err := misc.FirstNonEmpty(errorsBind, required_bindings); err != nil {
		return nil, err
	}

	wg.Add(len(componentsToSpawn[types.ComponentUPS]))
	errorsUPS := make(chan error, len(componentsToSpawn[types.ComponentUPS]))
	resultsUPS := make(chan types.ComponentClone, len(componentsToSpawn[types.ComponentUPS]))
	// Create dependent UPSes
	log.Infof("Creating dependent user provided services")
	required_bindings = 0
	for _, comp := range componentsToSpawn[types.ComponentUPS] {
		required_bindings += len(comp.DependencyOf)
		url := ""
		for _, app := range componentsToSpawn[types.ComponentApp] {
			for _, dep := range app.DependencyOf {
				if comp.GUID == dep {
					url = destAppsResources[app.GUID].Meta.URL
				}
			}
		}
		go c.createUserProvidedService(destApp.Entity.SpaceGUID, comp, suffix, url, &wg, resultsUPS, errorsUPS)
	}
	wg.Wait()
	close(errorsUPS)
	close(resultsUPS)
	if err := misc.FirstNonEmpty(errorsUPS, len(componentsToSpawn[types.ComponentUPS])); err != nil {
		return nil, err
	}
	log.Infof("Required bindings: %v", required_bindings)
	wg.Add(required_bindings)
	errorsBindUPS := make(chan error, required_bindings)
	// Bind UPSes
	log.Infof("Binding dependent user provided services")
	for clone := range resultsUPS {
		for _, dependent := range clone.Component.DependencyOf {
			go c.bindService(destAppsResources[dependent].Meta.GUID, clone.CloneGUID, &wg, errorsBindUPS)
		}
	}
	wg.Wait()
	close(errorsBindUPS)
	if err := misc.FirstNonEmpty(errorsBindUPS, required_bindings); err != nil {
		return nil, err
	}

	//Waiting for copy_bits finish
	log.Infof("Waiting for copy bits completion")
	if err := misc.FirstNonEmpty(copyBitsAsyncErrors, len(destAppsResources)); err != nil {
		return nil, err
	}

	// Starting applications one by one, not in parallel
	log.Infof("Starting applications")
	for _, comp := range componentsToSpawn[types.ComponentApp] {
		if err := c.startApp(destAppsResources[comp.GUID]); err != nil {
			return nil, err
		}
		log.Infof("Application %v started", destAppsResources[comp.GUID].Entity.Name)
	}

	log.Infof("Service instance [%v] created", destApp.Entity.Name)

	toReturn := types.ServiceCreationResponse{
		App: *destApp,
		ServiceCreationResponse: cf.ServiceCreationResponse{DashboardURL: ""},
	}
	return &toReturn, nil
}

// Deprovision remove instance of given application (that stands behind service instance though)
func (c *CfAPI) Deprovision(appGUID string) error {
	order, _ := c.DryRun(appGUID)
	log.Infof("Dry run: [%v]", order)
	log.Infof("%v components to remove:", len(order))

	componentsToRemove := make(map[types.ComponentType][]types.Component)
	componentsToRemove[types.ComponentApp] = []types.Component{}
	componentsToRemove[types.ComponentUPS] = []types.Component{}
	componentsToRemove[types.ComponentService] = []types.Component{}
	for _, comp := range order {
		if _, ok := componentsToRemove[comp.Type]; ok {
			componentsToRemove[comp.Type] = append(componentsToRemove[comp.Type], comp)
		} else {
			componentsToRemove[comp.Type] = []types.Component{comp}
		}
	}
	log.Infof("%+v", componentsToRemove)

	wg := sync.WaitGroup{}

	// Unbind services and UPSes
	log.Infof("Unbinding services and user provided services")
	results := make(chan error, len(componentsToRemove[types.ComponentApp]))
	wg.Add(len(componentsToRemove[types.ComponentApp]))
	for _, app := range componentsToRemove[types.ComponentApp] {
		go c.unbindAppServices(app.GUID, results, &wg)
	}
	wg.Wait()
	close(results)
	if err := misc.FirstNonEmpty(results, len(componentsToRemove[types.ComponentApp])); err != nil {
		if !c.isErrorAcceptedDuringDeprovision(err) {
			return err
		}
	}

	log.Infof("Removing service instances without bindings")
	resultsSvc := make(chan error, len(componentsToRemove[types.ComponentService]))
	wg.Add(len(componentsToRemove[types.ComponentService]))
	for _, svc := range componentsToRemove[types.ComponentService] {
		go c.deleteServiceInstIfUnbound(svc, resultsSvc, &wg)
	}
	wg.Wait()
	close(resultsSvc)
	if err := misc.FirstNonEmpty(resultsSvc, len(componentsToRemove[types.ComponentService])); err != nil {
		if !c.isErrorAcceptedDuringDeprovision(err) {
			return err
		}
	}

	log.Infof("Removing user provided service instances without bindings")
	resultsUPS := make(chan error, len(componentsToRemove[types.ComponentUPS]))
	wg.Add(len(componentsToRemove[types.ComponentUPS]))
	for _, ups := range componentsToRemove[types.ComponentUPS] {
		go c.deleteUPSInstIfUnbound(ups, resultsUPS, &wg)
	}
	wg.Wait()
	close(resultsUPS)
	if err := misc.FirstNonEmpty(resultsUPS, len(componentsToRemove[types.ComponentUPS])); err != nil {
		if !c.isErrorAcceptedDuringDeprovision(err) {
			return err
		}
	}

	log.Infof("Unbinding and deleting application routes")
	resultsRoutes := make(chan error, len(componentsToRemove[types.ComponentApp]))
	wg.Add(len(componentsToRemove[types.ComponentApp]))
	for _, app := range componentsToRemove[types.ComponentApp] {
		go c.deleteRoutes(app.GUID, resultsRoutes, &wg)
	}
	wg.Wait()
	close(resultsRoutes)
	if err := misc.FirstNonEmpty(resultsRoutes, len(componentsToRemove[types.ComponentApp])); err != nil {
		if !c.isErrorAcceptedDuringDeprovision(err) {
			return err
		}
	}

	log.Infof("Deleting applications")
	for _, app := range componentsToRemove[types.ComponentApp] {
		_ = c.deleteApp(app.GUID)
	}

	return nil
}

func (c *CfAPI) isErrorAcceptedDuringDeprovision(err error) bool {
	switch err {
	case misc.EntityNotFoundError{}, misc.InstanceNotFoundError{}, misc.ServiceNotFoundError{}:
		log.Errorf("Accepted error occured during deprovisioning: %v", err.Error())
		return true
	}
	return false
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

func (c *CfAPI) createService(spaceGUID string, comp types.Component, suffix string,
	wg *sync.WaitGroup, results chan types.ComponentClone, errorsCh chan error) {

	defer wg.Done()

	if len(comp.DependencyOf) == 0 {
		errorsCh <- errors.New("Service not attached to any application")
	}
	parentApp, err := c.getAppSummary(comp.DependencyOf[0])
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
	response, err := c.createServiceInstance(svcInstanceReq)
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

func (c *CfAPI) createApplication(sourceAppGUID, spaceGUID string, parameters map[string]string) (*types.CfAppResource, error) {
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
	destApp, err = c.createApp(destApp.Entity)
	if err != nil {
		return nil, err
	}

	domainGUID := sourceAppSummary.Routes[0].Domain.GUID
	domainName := sourceAppSummary.Routes[0].Domain.Name

	route, err := c.createRoute(&types.CfCreateRouteRequest{requestedName, domainGUID, spaceGUID})
	if err != nil {
		return nil, err
	}

	if err := c.associateRoute(destApp.Meta.GUID, route.Meta.GUID); err != nil {
		return nil, err
	}

	destApp.Meta.URL = fmt.Sprintf("%s.%s", route.Entity.Host, domainName)
	return destApp, nil
}

func (c *CfAPI) createUserProvidedService(spaceGUID string, comp types.Component, suffix, url string,
	wg *sync.WaitGroup, results chan types.ComponentClone, errorsCh chan error) {

	defer wg.Done()
	serviceName := comp.Name + "-" + suffix
	log.Debugf("Create dependent user provided service: service=[%v])", serviceName)

	// Retrieve UPS
	response, err := c.getUserProvidedService(comp.GUID)
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
	_ = c.applyAdditionalReplacementsInUPSCredentials(response)

	response, err = c.createUserProvidedServiceInstance(&response.Entity)
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

func (c *CfAPI) bindService(appGUID, serviceGUID string, wg *sync.WaitGroup, errors chan error) {
	defer wg.Done()
	// Bind created service
	svcBindingReq := types.NewCfServiceBindingRequest(appGUID, serviceGUID)
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
	log.Infof("Final UPS %v content %v", response.Entity.Name, credentialsStr)
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

func (c *CfAPI) deleteServiceInstIfUnbound(comp types.Component, errorsCh chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()

	bindings, err := c.getServiceBindings(comp.GUID)
	if err != nil {
		errorsCh <- err
		return
	}
	if bindings.TotalResults == 0 {
		log.Infof("Service %v is not bound to anything", comp.Name)
		log.Infof("Deleting %v instance %v", comp.Type, comp.Name)
		if err := c.deleteServiceInstance(comp.GUID); err != nil {
			errorsCh <- err
			return
		}
	} else {
		log.Infof("%v instance %v is bound to %v apps. Not deleting instance.", comp.Type, comp.Name, bindings.TotalResults)
	}
	errorsCh <- nil
}

func (c *CfAPI) deleteUPSInstIfUnbound(comp types.Component, errorsCh chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()

	bindings, err := c.getUserProvidedServiceBindings(comp.GUID)
	if err != nil {
		errorsCh <- err
		return
	}
	if bindings.TotalResults == 0 {
		log.Infof("Service %v is not bound to anything", comp.Name)
		log.Infof("Deleting %v instance %v", comp.Type, comp.Name)
		if err := c.deleteUserProvidedServiceInstance(comp.GUID); err != nil {
			errorsCh <- err
			return
		}
	} else {
		log.Infof("%v instance %v is bound to %v apps. Not deleting instance.", comp.Type, comp.Name, bindings.TotalResults)
	}
	errorsCh <- nil
}

func (c *CfAPI) unbindAppServices(appGUID string, result chan error, doneWaitGroup *sync.WaitGroup) {
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
			results <- nil
		}(loopBinding)
	}
	wg.Wait()
	result <- misc.FirstNonEmpty(results, len(bindings.Resources))
}

func (c *CfAPI) deleteRoutes(appGUID string, result chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()
	appSummary, err := c.getAppSummary(appGUID)
	if err != nil {
		result <- err
	}
	routes := appSummary.Routes

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
