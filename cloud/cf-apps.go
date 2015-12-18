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
	"time"
)

func (c *CfAPI) createApp(app types.CfApp) (*types.CfAppResource, error) {
	address := c.BaseAddress + "/v2/apps"
	log.Infof("Requesting app creation: %v", address)
	m, _ := json.Marshal(app)
	log.Debugf("Creating new app: [%+v]", app)
	resp, err := c.Post(address, "application/json", bytes.NewReader(m))

	if err != nil {
		log.Errorf("Could not create new app: [%v]", err)
		return nil, misc.InternalServerError{}
	}
	if resp.StatusCode != http.StatusCreated {
		log.Errorf("CreateApp failed. Response from CC: [%v]", misc.ReaderToString(resp.Body))
		return nil, misc.InternalServerError{}
	}

	toReturn := new(types.CfAppResource)
	json.NewDecoder(resp.Body).Decode(toReturn)
	log.Debugf("CreateApp status code: [%v]", resp.StatusCode)
	log.Debugf("App created. GUID: [%v]", toReturn.Meta.GUID)
	return toReturn, nil
}

func (c *CfAPI) getAppSummary(id string) (*types.CfAppSummary, error) {
	address := fmt.Sprintf("%v/v2/apps/%v/summary", c.BaseAddress, id)
	resp, err := c.getEntity(address, "application summary")
	if err != nil {
		return nil, err
	}

	toReturn := new(types.CfAppSummary)
	if err := json.NewDecoder(resp.Body).Decode(toReturn); err != nil {
		log.Errorf("Error decoding AppSummary response: [%v]", err);
		return nil, err
	}
	log.Debugf("getAppSummary status code: [%v]", resp.StatusCode)
	log.Debugf("AppSummary retrieved. [%+v]", toReturn)
	return toReturn, nil
}

func (c *CfAPI) deleteApp(id string) error {
	address := fmt.Sprintf("%v/v2/apps/%v", c.BaseAddress, id)
	return c.deleteEntity(address, "application")
}

func (c *CfAPI) getAppBindigs(id string) (*types.CfBindingsResources, error) {
	address := fmt.Sprintf("%v/v2/apps/%v/service_bindings", c.BaseAddress, id)
	response, err := c.getEntity(address, "app bindings")
	if err != nil {
		return nil, err
	}

	toReturn := new(types.CfBindingsResources)
	json.NewDecoder(response.Body).Decode(toReturn)
	log.Debugf("Get bindings status code: [%v]", response.StatusCode)
	log.Debugf("Bindings retrieved. Got %d of %d results", len(toReturn.Resources), toReturn.TotalResults)
	return toReturn, nil
}

func (c *CfAPI) deleteBinding(binding types.CfBindingResource) error {
	address := fmt.Sprintf("%v/v2/apps/%v/service_bindings/%v", c.BaseAddress, binding.Entity.AppGUID, binding.Meta.GUID)
	err := c.deleteEntity(address, "binding")
	if err != nil {
		log.Errorf("Error unbinding service instance %v from app %v", binding.Entity.ServiceInstanceGUID, binding.Entity.AppGUID)
		return err
	}
	return nil
}

func (c *CfAPI) copyBits(sourceID string, destID string, asyncError chan error) {
	address := fmt.Sprintf("%v/v2/apps/%v/copy_bits", c.BaseAddress, destID)
	log.Infof("Requesting copy_bits: %v", address)
	request := types.CfCopyBitsRequest{SrcAppGUID: sourceID}
	rawRequest, _ := json.Marshal(request)
	resp, err := c.Post(address, "application/json", bytes.NewReader(rawRequest))

	if err != nil {
		log.Errorf("Could not copy bits: [%v]", err)
		asyncError <- misc.InvalidInputError{}
		return
	} else if resp.StatusCode != http.StatusCreated {
		log.Errorf("CopyBits failed. Response from CC: [%v]", misc.ReaderToString(resp.Body))
		asyncError <- misc.InternalServerError{}
		return
	}

	jobResponse := new(types.CfJobResponse)
	json.NewDecoder(resp.Body).Decode(jobResponse)
	for jobResponse.Entity.Status != "finished" {
		if resp, err = c.Get(c.BaseAddress + jobResponse.Meta.URL); err != nil {
			asyncError <- misc.CcJobFailedError{err.Error()}
			return
		}
		json.NewDecoder(resp.Body).Decode(jobResponse)
		log.Debugf("Copy_bits job check: [%v]", jobResponse.Entity.Status)
		if jobResponse.Entity.Status == "failed" {
			asyncError <- misc.CcJobFailedError{jobResponse.Entity.Error}
			return
		}
		if jobResponse.Entity.Status == "queued" {
			time.Sleep(time.Second * 5)
		}
	}

	log.Debugf("CopyBits status code: [%v]", resp.StatusCode)
	log.Debugf("CopyBits finished")
	asyncError <- nil
}

func (c *CfAPI) restageApp(appGUID string) error {
	address := fmt.Sprintf("%v/v2/apps/%v/restage", c.BaseAddress, appGUID)
	log.Infof("Requesting restage: %v", address)

	request, err := http.NewRequest("POST", address, nil)
	resp, err := c.Do(request)
	if err != nil {
		log.Errorf("Could not restage app: [%v]", err)
		return misc.CcRestageFailedError{err.Error()}
	} else if resp.StatusCode != http.StatusCreated {
		log.Errorf("RestageApp finished with error: %v", misc.ReaderToString(resp.Body))
		return misc.CcRestageFailedError{"Unexpected HTTP status returned from CC"}
	}

	restagedApp := new(types.CfAppResource)
	json.NewDecoder(resp.Body).Decode(restagedApp)
	log.Debugf("RestageApp status code: [%v]", resp.StatusCode)
	log.Debugf("App status after restage: [%v]", restagedApp.Entity.State)
	return nil
}

func (c *CfAPI) updateApp(app *types.CfAppResource) error {
	address := fmt.Sprintf("%v/v2/apps/%v", c.BaseAddress, app.Meta.GUID)
	log.Infof("Updating an app: %v", address)
	raw, _ := json.Marshal(app.Entity)
	request, _ := http.NewRequest("PUT", address, bytes.NewReader(raw))
	resp, err := c.Do(request)
	if err != nil {
		log.Errorf("Could not update app: [%v]", err)
		return misc.CcUpdateFailedError{err.Error()}
	} else if resp.StatusCode != http.StatusCreated {
		log.Errorf("UpdateApp finished with error: %v", misc.ReaderToString(resp.Body))
		return misc.CcUpdateFailedError{"Unexpected HTTP status returned from CC:" + resp.Status}
	}
	return nil
}

func (c *CfAPI) startApp(app *types.CfAppResource) error {
	app.Entity.State = types.AppStarted
	if err := c.updateApp(app); err != nil {
		return err
	}

	var err error
	asyncErr := make(chan error)
	go c.waitForAppRunning(app.Meta.GUID, asyncErr)
	select {
	case err = <-asyncErr:
		if err != nil {
			return err
		}
	case <-time.After(5 * time.Minute):
		return misc.TimeoutOccurredError{}
	}
	return nil
}

func (c *CfAPI) waitForAppRunning(appGUID string, asyncErr chan error) {
	address := fmt.Sprintf("%v/v2/apps/%v/instances", c.BaseAddress, appGUID)
	log.Infof("Waiting for app running, checking instances: %v", address)

	timeout := time.Second * 5
	for {
		resp, err := c.Get(address)
		if err != nil {
			log.Errorf("Could not get app instances: [%v]", err)
			asyncErr <- misc.CcGetInstancesFailedError{err.Error()}
			return
		} else if resp.StatusCode != http.StatusOK {
			log.Debugf("waitForAppRunning finished with error: %v", misc.ReaderToString(resp.Body))
			time.Sleep(timeout)
			continue
		}

		decodedInstances := map[string]types.CfAppInstance{}
		if err := json.NewDecoder(resp.Body).Decode(&decodedInstances); err != nil {
			asyncErr <- misc.CcGetInstancesFailedError{err.Error()}
			return
		}

		running := true
		for key, value := range decodedInstances {
			log.Infof("Instance %v, status: %v", key, value)
			if value.State != "RUNNING" {
				running = false
				time.Sleep(timeout)
				break
			}
		}
		if running {
			break
		}
	}
	asyncErr <- nil
	return
}
