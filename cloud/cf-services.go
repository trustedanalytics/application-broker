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
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"net/http"
)

func (c *CfAPI) createServiceInstance(req *types.CfServiceInstanceCreateRequest) (*types.CfServiceInstanceCreateResponse, error) {
	address := c.BaseAddress + "/v2/service_instances?accepts_incomplete=false"
	log.Infof("Requesting service instance creation: %v", address)
	marshalled, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Could not marshal CfServiceInstanceCreateRequest: [%+v]", req)
		return nil, misc.InternalServerError{Context: "Problem with marshalling request data"}
	}
	resp, err := c.Post(address, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		log.Errorf("Could not create service instance: [%v]", err)
		return nil, misc.InternalServerError{Context: "Cloud Foundry API was not able to create service instance"}
	}
	if !(resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted) {
		// CF 2.07 returns HTTP 201, CF 2.22 returns HTTP 202
		log.Errorf("createServiceInstance failed. Response from CC: [%v]", misc.ReaderToString(resp.Body))
		return nil, misc.InternalServerError{Context: "Unacceptable response code from Cloud Foundry API after trying to create service instance"}
	}

	toReturn := new(types.CfServiceInstanceCreateResponse)
	json.NewDecoder(resp.Body).Decode(toReturn)
	log.Debugf("createServiceInstance status code: [%v]", resp.StatusCode)
	log.Debugf("createServiceInstance returned GUID: [%v]", toReturn.Meta.GUID)
	return toReturn, nil
}

func (c *CfAPI) createServiceBinding(req *types.CfServiceBindingCreateRequest) (*types.CfServiceBindingCreateResponse, error) {
	address := c.BaseAddress + "/v2/service_bindings"
	log.Infof("Requesting service binding creation: %v", address)
	marshalled, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Could not marshal CfServiceInstanceCreateRequest: [%+v]", req)
		return nil, misc.InternalServerError{Context: "Problem with marshalling request data"}
	}
	resp, err := c.Post(address, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		log.Errorf("Could not create service binding: [%v]", err)
		return nil, misc.InternalServerError{Context: "Cloud Foundry API was not able to create service binding"}
	}
	if resp.StatusCode != http.StatusCreated {
		log.Errorf("createServiceBinding failed. Response from CC: [%v]", misc.ReaderToString(resp.Body))
		return nil, misc.InternalServerError{Context: "Unacceptable response code from Cloud Foundry API after trying to create service binding"}
	}

	toReturn := new(types.CfServiceBindingCreateResponse)
	json.NewDecoder(resp.Body).Decode(toReturn)
	log.Debugf("createServiceBinding status code: [%v]", resp.StatusCode)
	log.Debugf("createServiceBinding returned GUID: [%v]", toReturn.Meta.GUID)
	return toReturn, nil
}

func (c *CfAPI) deleteServiceInstance(id string) error {
	address := fmt.Sprintf("%v/v2/service_instances/%v", c.BaseAddress, id)
	err := c.deleteEntity(address, "service instance")
	if err != nil {
		log.Errorf("Error deleting service instance %v", id)
		return err
	}
	return nil
}
