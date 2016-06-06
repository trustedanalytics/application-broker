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
	"github.com/signalfx/golib/errors"
	"github.com/trustedanalytics/application-broker/client"
	"github.com/trustedanalytics/application-broker/env"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/service/extension"
	"github.com/trustedanalytics/go-cf-lib/api"
	"github.com/trustedanalytics/go-cf-lib/types"
	"strings"
	"sync"
)

type CloudAPI struct {
	cf            *api.CfAPI
	appDepDiscUps *client.AppDependencyDiscovererUPS
}

func NewCloudAPI(envs *cfenv.App) *CloudAPI {
	toReturn := new(CloudAPI)
	toReturn.cf = api.NewCfAPI()
	toReturn.appDepDiscUps = client.NewAppDependencyDiscovererUPS(envs)

	return toReturn
}

// Provision instantiates service of given type
func (cloud *CloudAPI) Provision(sourceAppGUID string,
	servicesConfiguration []*extension.ServiceConfiguration,
	r *cf.ServiceCreationRequest) (*extension.ServiceCreationResponse, error) {

	order, _ := cloud.Discovery(sourceAppGUID)
	log.Infof("Discovery: [%v]", order)
	log.Infof("%v components to spawn:", len(order))

	cloud.logParameters(r.Parameters, servicesConfiguration)

	componentsToSpawn := cloud.groupComponentsByType(order)

	suffix := strings.Split(r.InstanceID, "-")[0]

	destAppsResources := make(map[string]*types.CfAppResource)
	transaction := NewTransaction()

	log.Infof("Creating main application")
	paramsWithoutNS, err := cloud.removeParametersNamespaces(r.Parameters)
	if err != nil {
		return nil, err
	}
	destApp, err := cloud.cf.CreateApplicationClone(sourceAppGUID, r.SpaceGUID, paramsWithoutNS)
	if err != nil {
		return nil, err
	}
	destAppsResources[sourceAppGUID] = destApp
	transaction.AddApplication(destApp)

	log.Infof("Creating dependent applications")
	for _, app := range componentsToSpawn[types.ComponentApp] {
		if _, ok := destAppsResources[app.GUID]; !ok {
			name := fmt.Sprintf("%v-%v", app.Name, suffix)
			paramsWithoutNS["name"] = name
			appRes, err := cloud.cf.CreateApplicationClone(app.GUID, r.SpaceGUID, paramsWithoutNS)
			if err != nil {
				transaction.Rollback(cloud)
				return nil, err
			}
			destAppsResources[app.GUID] = appRes
			transaction.AddApplication(appRes)
		}
	}

	log.Infof("Copying applications data")
	copyBitsAsyncErrors := make(chan error, len(destAppsResources))
	for _, appRes := range destAppsResources {
		go cloud.cf.CopyBits(sourceAppGUID, appRes.Meta.GUID, copyBitsAsyncErrors)
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
		go cloud.cf.CreateServiceClone(destApp.Entity.SpaceGUID,
			cloud.selectAcceptedServiceParams(comp.Name, r.Parameters, servicesConfiguration),
			comp, suffix, results, errors, &wg)
	}
	wg.Wait()
	close(errors)
	close(results)
	if err := misc.FirstNonEmpty(errors, len(componentsToSpawn[types.ComponentService])); err != nil {
		for clone := range results {
			transaction.AddComponentClone(&clone)
		}
		transaction.Rollback(cloud)
		return nil, err
	}
	log.Infof("Required bindings: %v", required_bindings)
	wg.Add(required_bindings)
	errorsBind := make(chan error, required_bindings)
	// Bind services
	log.Infof("Binding dependent services")
	for clone := range results {
		for _, dependent := range clone.Component.DependencyOf {
			go cloud.cf.BindService(destAppsResources[dependent].Meta.GUID, clone.CloneGUID, errorsBind, &wg)
		}
	}
	wg.Wait()
	close(errorsBind)
	if err := misc.FirstNonEmpty(errorsBind, required_bindings); err != nil {
		transaction.Rollback(cloud)
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
		go cloud.CreateUserProvidedServiceClone(destApp.Entity.SpaceGUID, comp, suffix, url,
			resultsUPS, errorsUPS, &wg)
	}
	wg.Wait()
	close(errorsUPS)
	close(resultsUPS)
	if err := misc.FirstNonEmpty(errorsUPS, len(componentsToSpawn[types.ComponentUPS])); err != nil {
		for clone := range resultsUPS {
			transaction.AddComponentClone(&clone)
		}
		transaction.Rollback(cloud)
		return nil, err
	}
	log.Infof("Required bindings: %v", required_bindings)
	wg.Add(required_bindings)
	errorsBindUPS := make(chan error, required_bindings)
	// Bind UPSes
	log.Infof("Binding dependent user provided services")
	for clone := range resultsUPS {
		for _, dependent := range clone.Component.DependencyOf {
			go cloud.cf.BindService(destAppsResources[dependent].Meta.GUID, clone.CloneGUID, errorsBindUPS, &wg)
		}
	}
	wg.Wait()
	close(errorsBindUPS)
	if err := misc.FirstNonEmpty(errorsBindUPS, required_bindings); err != nil {
		transaction.Rollback(cloud)
		return nil, err
	}

	//Waiting for copy_bits finish
	log.Infof("Waiting for copy bits completion")
	if err := misc.FirstNonEmpty(copyBitsAsyncErrors, len(destAppsResources)); err != nil {
		transaction.Rollback(cloud)
		return nil, err
	}

	// Starting applications one by one, not in parallel
	log.Infof("Starting applications")
	for _, comp := range componentsToSpawn[types.ComponentApp] {
		if err := cloud.cf.StartApp(destAppsResources[comp.GUID]); err != nil {
			transaction.Rollback(cloud)
			return nil, err
		}
		log.Infof("Application %v started", destAppsResources[comp.GUID].Entity.Name)
	}

	log.Infof("Service instance [%v] created", destApp.Entity.Name)

	toReturn := extension.ServiceCreationResponse{
		App: *destApp,
		ServiceCreationResponse: cf.ServiceCreationResponse{DashboardURL: ""},
	}
	return &toReturn, nil
}

// Deprovision remove instance of given application (that stands behind service instance though)
func (cloud *CloudAPI) Deprovision(appGUID string) error {
	order, _ := cloud.Discovery(appGUID)
	log.Infof("Discovery: [%v]", order)
	log.Infof("%v components to remove:", len(order))

	return cloud.deprovisionComponents(order)
}

func (cloud *CloudAPI) deprovisionComponents(order []types.Component) error {
	componentsToRemove := cloud.groupComponentsByType(order)

	wg := sync.WaitGroup{}

	// Unbind services and UPSes
	log.Infof("Unbinding services and user provided services")
	results := make(chan error, len(componentsToRemove[types.ComponentApp]))
	wg.Add(len(componentsToRemove[types.ComponentApp]))
	for _, app := range componentsToRemove[types.ComponentApp] {
		go cloud.cf.UnbindAppServices(app.GUID, results, &wg)
	}
	wg.Wait()
	close(results)
	for err := range results {
		if !cloud.isErrorAcceptedDuringDeprovision(err) {
			log.Errorf("Error occured when unbinding services and upses: %v", err.Error())
		}
	}

	log.Infof("Removing service instances without bindings")
	resultsSvc := make(chan error, len(componentsToRemove[types.ComponentService]))
	wg.Add(len(componentsToRemove[types.ComponentService]))
	for _, svc := range componentsToRemove[types.ComponentService] {
		go cloud.cf.DeleteServiceInstIfUnbound(svc, resultsSvc, &wg)
	}
	wg.Wait()
	close(resultsSvc)
	for err := range resultsSvc {
		if !cloud.isErrorAcceptedDuringDeprovision(err) {
			log.Errorf("Error occured when removing service instances: %v", err.Error())
		}
	}

	log.Infof("Removing user provided service instances without bindings")
	resultsUPS := make(chan error, len(componentsToRemove[types.ComponentUPS]))
	wg.Add(len(componentsToRemove[types.ComponentUPS]))
	for _, ups := range componentsToRemove[types.ComponentUPS] {
		go cloud.cf.DeleteUPSInstIfUnbound(ups, resultsUPS, &wg)
	}
	wg.Wait()
	close(resultsUPS)
	for err := range resultsUPS {
		if !cloud.isErrorAcceptedDuringDeprovision(err) {
			log.Errorf("Error occured when removing user provided service instances: %v", err.Error())
		}
	}

	log.Infof("Unbinding and deleting application routes")
	resultsRoutes := make(chan error, len(componentsToRemove[types.ComponentApp]))
	wg.Add(len(componentsToRemove[types.ComponentApp]))
	for _, app := range componentsToRemove[types.ComponentApp] {
		go cloud.cf.DeleteRoutes(app.GUID, resultsRoutes, &wg)
	}
	wg.Wait()
	close(resultsRoutes)
	for err := range resultsRoutes {
		if !cloud.isErrorAcceptedDuringDeprovision(err) {
			log.Errorf("Error occured when unbinding and deleting application routes: %v", err.Error())
		}
	}

	log.Infof("Deleting applications")
	for _, app := range componentsToRemove[types.ComponentApp] {
		_ = cloud.cf.DeleteApp(app.GUID)
	}

	return nil
}

// UpdateBroker registers or updates catalog in CF
func (cloud *CloudAPI) UpdateBroker(brokerName string, brokerURL string, username string, password string) error {
	brokers, err := cloud.cf.GetBrokers(brokerName)
	if err != nil {
		return err
	}

	if brokers.TotalResults == 0 {
		return cloud.cf.RegisterBroker(brokerName, brokerURL, username, password)
	}
	return cloud.cf.UpdateBroker(brokers.Resources[0].Meta.GUID, brokerURL, username, password)
}

func (cloud *CloudAPI) CheckIfServiceExists(serviceName string) error {
	myData := env.GetVcapApplication()
	broker, err := cloud.cf.GetBrokers(myData.Name)
	duplicate, err := cloud.cf.GetServiceOfName(serviceName)
	if err != nil {
		return err
	}
	if duplicate != nil {
		if broker.TotalResults == 0 || broker.Resources[0].Meta.GUID != duplicate.Entity.BrokerGUID {
			return errors.Annotate(types.InternalServerError, "Service name already registered in different CF broker!")
		} else if broker.TotalResults > 0 && broker.Resources[0].Meta.GUID == duplicate.Entity.BrokerGUID {
			log.Infof("Service name was registered in CF for THIS broker but was missing in internal DB, purging...", serviceName)
			return cloud.cf.PurgeService(duplicate.Meta.GUID, duplicate.Entity.Name, duplicate.Entity.PlansURL)
		}
	}
	return nil
}
