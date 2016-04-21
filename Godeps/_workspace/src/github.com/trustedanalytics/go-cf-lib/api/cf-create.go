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

package api

import (
	"errors"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/go-cf-lib/types"
	"sync"
)

func (c *CfAPI) CreateServiceClone(spaceGUID string, comp types.Component, suffix string,
	resultsCh chan types.ComponentClone, errorsCh chan error, wg *sync.WaitGroup) {

	defer wg.Done()

	if len(comp.DependencyOf) == 0 {
		errorsCh <- errors.New("Service not attached to any application")
		return
	}
	parentApp, err := c.GetAppSummary(comp.DependencyOf[0])
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
	log.Debugf("Create dependent service: service=[%v] ([%v], [%v])",
		serviceName, svc.Plan.Service.Label, svc.Plan.Name)

	// Create service
	svcInstanceReq := types.NewCfServiceInstanceRequest(serviceName, spaceGUID, svc.Plan)
	response, err := c.CreateServiceInstance(svcInstanceReq)
	if err != nil {
		errorsCh <- err
		return
	}
	spawnedServiceInstanceGUID := response.Meta.GUID
	log.Debugf("Dependent service created: Service Instance GUID=[%v]", spawnedServiceInstanceGUID)

	resultsCh <- types.ComponentClone{
		Component: comp,
		CloneGUID: spawnedServiceInstanceGUID,
	}
	errorsCh <- nil
	return
}

func (c *CfAPI) CreateApplicationClone(sourceAppGUID, spaceGUID string, parameters map[string]string) (*types.CfAppResource, error) {
	// Gather reference app summary to be used later for creating new instance
	sourceAppSummary, err := c.GetAppSummary(sourceAppGUID)
	if err != nil {
		return nil, err
	}
	requestedName := parameters["name"]
	delete(parameters, "name")

	err = c.AssertAppHasRoutes(sourceAppSummary)
	if err != nil {
		return nil, err
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
	destApp, err = c.CreateApp(destApp.Entity)
	if err != nil {
		return nil, err
	}

	domainGUID := sourceAppSummary.Routes[0].Domain.GUID
	domainName := sourceAppSummary.Routes[0].Domain.Name

	route, err := c.CreateRoute(&types.CfCreateRouteRequest{requestedName, domainGUID, spaceGUID})
	if err != nil {
		return nil, err
	}

	if err := c.AssociateRoute(destApp.Meta.GUID, route.Meta.GUID); err != nil {
		return nil, err
	}

	destApp.Meta.URL = fmt.Sprintf("%s.%s", route.Entity.Host, domainName)
	return destApp, nil
}
