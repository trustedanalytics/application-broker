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
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/signalfx/golib/errors"
	"github.com/trustedanalytics/go-cf-lib/helpers"
	"github.com/trustedanalytics/go-cf-lib/types"
	"net/http"
)

func (c *CfAPI) CreateUserProvidedServiceInstance(req *types.CfUserProvidedService) (*types.CfUserProvidedServiceResource, error) {
	address := c.BaseAddress + "/v2/user_provided_service_instances"
	log.Infof("Requesting user provided service instance creation: %v", address)
	marshalled, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Could not marshal CfUserProvidedService: [%+v]", req)
		return nil, errors.Annotate(types.InternalServerError, "Problem with marshalling request data")
	}
	resp, err := c.Post(address, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		log.Errorf("Could not create user provided service instance: [%v]", err)
		return nil, errors.Annotate(types.InternalServerError, "Cloud Foundry API was not able to create user provided service instance")
	}
	if !(resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted) {
		// CF 2.07 returns HTTP 201, CF 2.22 returns HTTP 202
		log.Errorf("createUserProvidedServiceInstance failed. Response from CC: [%v]", helpers.ReaderToString(resp.Body))
		return nil, errors.Annotate(types.InternalServerError, "Unacceptable response code from Cloud Foundry API after trying to create service instance")
	}

	toReturn := new(types.CfUserProvidedServiceResource)
	json.NewDecoder(resp.Body).Decode(toReturn)
	log.Debugf("createUserProvidedServiceInstance status code: [%v]", resp.StatusCode)
	log.Debugf("createUserProvidedServiceInstance returned GUID: [%v]", toReturn.Meta.GUID)
	return toReturn, nil
}

func (c *CfAPI) GetUserProvidedService(guid string) (*types.CfUserProvidedServiceResource, error) {
	address := fmt.Sprintf("%v/v2/user_provided_service_instances/%v", c.BaseAddress, guid)
	log.Infof("Requesting user provided service retrieval: %v", address)
	resp, err := c.getEntity(address, "user provided service")
	if err != nil {
		return nil, err
	}

	toReturn := new(types.CfUserProvidedServiceResource)
	err = json.NewDecoder(resp.Body).Decode(toReturn)
	if err != nil {
		return nil, err
	}
	log.Debugf("User provided service with guid [%v] found", guid)
	return toReturn, nil
}

func (c *CfAPI) CreateUserProvidedServiceBinding(req *types.CfServiceBindingCreateRequest) (*types.CfServiceBindingCreateResponse, error) {
	address := c.BaseAddress + "/v2/service_bindings"
	log.Infof("Requesting service binding creation: %v", address)
	marshalled, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Could not marshal CfServiceInstanceCreateRequest: [%+v]", req)
		return nil, errors.Annotate(types.InternalServerError, "Problem with marshalling request data")
	}
	resp, err := c.Post(address, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		log.Errorf("Could not create service binding: [%v]", err)
		return nil, errors.Annotate(types.InternalServerError, "Cloud Foundry API was not able to create service binding")
	}
	if resp.StatusCode != http.StatusCreated {
		log.Errorf("createServiceBinding failed. Response from CC: [%v]", helpers.ReaderToString(resp.Body))
		return nil, errors.Annotate(types.InternalServerError, "Unacceptable response code from Cloud Foundry API after trying to create service binding")
	}

	toReturn := new(types.CfServiceBindingCreateResponse)
	json.NewDecoder(resp.Body).Decode(toReturn)
	log.Debugf("createServiceBinding status code: [%v]", resp.StatusCode)
	log.Debugf("createServiceBinding returned GUID: [%v]", toReturn.Meta.GUID)
	return toReturn, nil
}

func (c *CfAPI) DeleteUserProvidedServiceInstance(id string) error {
	address := fmt.Sprintf("%v/v2/user_provided_service_instances/%v", c.BaseAddress, id)
	err := c.deleteEntity(address, "UPS instance")
	if err != nil {
		log.Errorf("Error deleting service instance %v", id)
		return err
	}
	return nil
}

func (c *CfAPI) GetUserProvidedServiceBindings(id string) (*types.CfBindingsResources, error) {
	address := fmt.Sprintf("%v/v2/user_provided_service_instances/%v/service_bindings", c.BaseAddress, id)
	response, err := c.getEntity(address, "service bindings")
	if err != nil {
		return nil, err
	}

	toReturn := new(types.CfBindingsResources)
	json.NewDecoder(response.Body).Decode(toReturn)
	log.Debugf("Get bindings status code: [%v]", response.StatusCode)
	log.Debugf("Bindings retrieved. Got %d of %d results", len(toReturn.Resources), toReturn.TotalResults)
	return toReturn, nil
}
