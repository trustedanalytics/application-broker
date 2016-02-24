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
	"github.com/cloudfoundry-community/types-cf"
	"github.com/nu7hatch/gouuid"
	"github.com/trustedanalytics/application-broker/cf-rest-api"
	"github.com/trustedanalytics/application-broker/graph"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"sync"
)

type CloudAPI struct {
	cf *api.CfAPI
}

func NewCloudAPI() *CloudAPI {
	toReturn := new(CloudAPI)
	toReturn.cf = api.NewCfAPI()
	return toReturn
}

// Returns a list of services and apps which would be provisioned in normal run
func (cl *CloudAPI) DryRun(sourceAppGUID string) ([]types.Component, error) {
	g := graph.NewGraphAPI()
	ret, err := g.DryRun(sourceAppGUID)
	return ret, err
}

// Provision instantiates service of given type
func (cl *CloudAPI) Provision(sourceAppGUID string, r *cf.ServiceCreationRequest) (*types.ServiceCreationResponse, error) {
	order, _ := cl.DryRun(sourceAppGUID)
	log.Infof("Dry run: [%v]", order)
	log.Infof("%v components to spawn:", len(order))

	componentsToSpawn := cl.groupComponentsByType(order)

	commonUUID, _ := uuid.NewV4()
	suffix := commonUUID.String()[:8]

	destAppsResources := make(map[string]*types.CfAppResource)

	log.Infof("Creating main application")
	destApp, err := cl.createApplication(sourceAppGUID, r.SpaceGUID, r.Parameters)
	if err != nil {
		return nil, err
	}
	destAppsResources[sourceAppGUID] = destApp
	log.Infof("Creating dependent applications")
	for _, app := range componentsToSpawn[types.ComponentApp] {
		if _, ok := destAppsResources[app.GUID]; !ok {
			name := fmt.Sprintf("%v-%v", app.Name, suffix)
			params := map[string]string{"name": name}
			appRes, err := cl.createApplication(app.GUID, r.SpaceGUID, params)
			if err != nil {
				return nil, err
			}
			destAppsResources[app.GUID] = appRes
		}
	}

	log.Infof("Copying applications data")
	copyBitsAsyncErrors := make(chan error, len(destAppsResources))
	for _, appRes := range destAppsResources {
		go cl.cf.CopyBits(sourceAppGUID, appRes.Meta.GUID, copyBitsAsyncErrors)
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
		go cl.createService(destApp.Entity.SpaceGUID, comp, suffix, &wg, results, errors)
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
			go cl.bindService(destAppsResources[dependent].Meta.GUID, clone.CloneGUID, &wg, errorsBind)
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
		go cl.createUserProvidedService(destApp.Entity.SpaceGUID, comp, suffix, url, &wg, resultsUPS, errorsUPS)
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
			go cl.bindService(destAppsResources[dependent].Meta.GUID, clone.CloneGUID, &wg, errorsBindUPS)
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
		if err := cl.cf.StartApp(destAppsResources[comp.GUID]); err != nil {
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
func (cl *CloudAPI) Deprovision(appGUID string) error {
	order, _ := cl.DryRun(appGUID)
	log.Infof("Dry run: [%v]", order)
	log.Infof("%v components to remove:", len(order))

	componentsToRemove := cl.groupComponentsByType(order)

	wg := sync.WaitGroup{}

	// Unbind services and UPSes
	log.Infof("Unbinding services and user provided services")
	results := make(chan error, len(componentsToRemove[types.ComponentApp]))
	wg.Add(len(componentsToRemove[types.ComponentApp]))
	for _, app := range componentsToRemove[types.ComponentApp] {
		go cl.unbindAppServices(app.GUID, results, &wg)
	}
	wg.Wait()
	close(results)
	if err := misc.FirstNonEmpty(results, len(componentsToRemove[types.ComponentApp])); err != nil {
		if !cl.isErrorAcceptedDuringDeprovision(err) {
			return err
		}
	}

	log.Infof("Removing service instances without bindings")
	resultsSvc := make(chan error, len(componentsToRemove[types.ComponentService]))
	wg.Add(len(componentsToRemove[types.ComponentService]))
	for _, svc := range componentsToRemove[types.ComponentService] {
		go cl.deleteServiceInstIfUnbound(svc, resultsSvc, &wg)
	}
	wg.Wait()
	close(resultsSvc)
	if err := misc.FirstNonEmpty(resultsSvc, len(componentsToRemove[types.ComponentService])); err != nil {
		if !cl.isErrorAcceptedDuringDeprovision(err) {
			return err
		}
	}

	log.Infof("Removing user provided service instances without bindings")
	resultsUPS := make(chan error, len(componentsToRemove[types.ComponentUPS]))
	wg.Add(len(componentsToRemove[types.ComponentUPS]))
	for _, ups := range componentsToRemove[types.ComponentUPS] {
		go cl.deleteUPSInstIfUnbound(ups, resultsUPS, &wg)
	}
	wg.Wait()
	close(resultsUPS)
	if err := misc.FirstNonEmpty(resultsUPS, len(componentsToRemove[types.ComponentUPS])); err != nil {
		if !cl.isErrorAcceptedDuringDeprovision(err) {
			return err
		}
	}

	log.Infof("Unbinding and deleting application routes")
	resultsRoutes := make(chan error, len(componentsToRemove[types.ComponentApp]))
	wg.Add(len(componentsToRemove[types.ComponentApp]))
	for _, app := range componentsToRemove[types.ComponentApp] {
		go cl.deleteRoutes(app.GUID, resultsRoutes, &wg)
	}
	wg.Wait()
	close(resultsRoutes)
	if err := misc.FirstNonEmpty(resultsRoutes, len(componentsToRemove[types.ComponentApp])); err != nil {
		if !cl.isErrorAcceptedDuringDeprovision(err) {
			return err
		}
	}

	log.Infof("Deleting applications")
	for _, app := range componentsToRemove[types.ComponentApp] {
		_ = cl.cf.DeleteApp(app.GUID)
	}

	return nil
}

// UpdateBroker registers or updates catalog in CF
func (cl *CloudAPI) UpdateBroker(brokerName string, brokerURL string, username string, password string) error {
	brokers, err := cl.cf.GetBrokers(brokerName)
	if err != nil {
		return err
	}

	if brokers.TotalResults == 0 {
		return cl.cf.RegisterBroker(brokerName, brokerURL, username, password)
	}
	return cl.cf.UpdateBroker(brokers.Resources[0].Meta.GUID, brokerURL, username, password)
}

func (cl *CloudAPI) CheckIfServiceExists(serviceName string) error {
	myData := types.GetVcapApplication()
	broker, err := cl.cf.GetBrokers(myData.Name)
	duplicate, err := cl.cf.GetServiceOfName(serviceName)
	if err != nil {
		return err
	}
	if duplicate != nil {
		if broker.TotalResults == 0 || broker.Resources[0].Meta.GUID != duplicate.Entity.BrokerGUID {
			return misc.InternalServerError{Context: "Service name already registered in different CF broker!"}
		} else if broker.TotalResults > 0 && broker.Resources[0].Meta.GUID == duplicate.Entity.BrokerGUID {
			log.Infof("Service name was registered in CF for THIS broker but was missing in internal DB, purging...", serviceName)
			return cl.cf.PurgeService(duplicate.Meta.GUID, duplicate.Entity.Name, duplicate.Entity.PlansURL)
		}
	}
	return nil
}
