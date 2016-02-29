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
	"github.com/trustedanalytics/go-cf-lib/helpers"
	"github.com/trustedanalytics/go-cf-lib/types"
	"net/http"
)

func (c *CfAPI) RegisterBroker(brokerName string, brokerURL string, username string, password string) error {
	address := fmt.Sprintf("%v/v2/service_brokers", c.BaseAddress)

	req := types.CfServiceBroker{Name: brokerName, URL: brokerURL, Username: username, Password: password}
	serialized, _ := json.Marshal(req)
	log.Infof("Registering broker: %v %v", address, serialized)

	request, err := http.NewRequest(MethodPost, address, bytes.NewReader(serialized))
	if err != nil {
		msg := fmt.Sprintf("Failed to prepare request for: %v %v", MethodPost, address)
		log.Error(msg)
		return types.InternalServerError{Context: msg}
	}
	response, err := c.Do(request)

	if err != nil {
		msg := fmt.Sprintf("Failed to register service broker: %v", err.Error())
		log.Error(msg)
		return types.InternalServerError{Context: msg}
	}

	if response.StatusCode != http.StatusCreated {
		msg := fmt.Sprintf("Failed to register service broker: Status code %d, Error %v", response.StatusCode,
			helpers.ReaderToString(response.Body))
		log.Error(msg)
		return types.InternalServerError{Context: msg}
	}

	return nil
}

func (c *CfAPI) UpdateBroker(brokerGUID string, brokerURL string, username string, password string) error {
	address := fmt.Sprintf("%v/v2/service_brokers/%v", c.BaseAddress, brokerGUID)

	req := types.CfServiceBroker{URL: brokerURL, Username: username, Password: password}
	serialized, _ := json.Marshal(req)

	log.Infof("Updating: %v %v", address, brokerURL)

	request, err := http.NewRequest(MethodPut, address, bytes.NewReader(serialized))
	if err != nil {
		msg := fmt.Sprintf("Failed to prepare request for: %v %v", MethodPost, address)
		log.Error(msg)
		return types.InternalServerError{Context: msg}
	}
	response, err := c.Do(request)

	if err != nil {
		msg := fmt.Sprintf("Failed to update service broker: %v", err.Error())
		log.Error(msg)
		return types.InternalServerError{Context: msg}
	}

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to update service broker: Status code %d, Error %v", response.StatusCode,
			helpers.ReaderToString(response.Body))
		log.Error(msg)
		return types.InternalServerError{Context: msg}
	}

	return nil
}

func (c *CfAPI) GetBrokers(brokerName string) (*types.CfServiceBrokerResources, error) {
	address := fmt.Sprintf("%v/v2/service_brokers?q=name:%v", c.BaseAddress, brokerName)
	response, err := c.Get(address)
	if err != nil {
		msg := fmt.Sprintf("Failed to get available service brokers: %v", err.Error())
		log.Error(msg)
		return nil, types.InternalServerError{Context: msg}
	}

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to get available service brokers: Status code %d, Error %v",
			response.StatusCode, helpers.ReaderToString(response.Body))
		log.Error(msg)
		return nil, types.InternalServerError{Context: msg}
	}

	brokers := new(types.CfServiceBrokerResources)
	if err := json.NewDecoder(response.Body).Decode(brokers); err != nil {
		msg := fmt.Sprintf("Failed to parse broker list response: %v", err.Error())
		log.Error(msg)
		return nil, types.InternalServerError{Context: msg}
	}
	return brokers, nil
}
