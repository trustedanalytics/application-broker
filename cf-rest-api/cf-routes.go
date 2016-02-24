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
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"net/http"
)

func (c *CfAPI) CreateRoute(req *types.CfCreateRouteRequest) (*types.CfRouteResource, error) {
	address := c.BaseAddress + "/v2/routes"
	log.Infof("Requesting route creation: %v", address)
	marshalled, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Could not marshal CfCreateRouteRequest: [%+v]", req)
		return nil, misc.InternalServerError{}
	}
	resp, err := c.Post(address, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		log.Errorf("Could not create route: [%v]", err)
		return nil, misc.InternalServerError{}
	}
	if resp.StatusCode != http.StatusCreated {
		log.Errorf("CreateRoute failed. Response from CC: [%v]", misc.ReaderToString(resp.Body))
		return nil, misc.InternalServerError{}
	}

	toReturn := new(types.CfRouteResource)
	json.NewDecoder(resp.Body).Decode(toReturn)
	log.Debugf("CreateRoute status code: [%v]", resp.StatusCode)
	log.Debugf("CreateRoute returned GUID: [%v]", toReturn.Meta.GUID)
	return toReturn, nil
}

func (c *CfAPI) AssociateRoute(appID string, routeID string) error {
	address := fmt.Sprintf("%v/v2/apps/%v/routes/%v", c.BaseAddress, appID, routeID)
	log.Infof("Requesting route association: %v", address)
	req, _ := http.NewRequest("PUT", address, nil)

	resp, err := c.Do(req)
	if err != nil {
		log.Errorf("Could not associate app with route: [%v]", err)
		return misc.InternalServerError{}
	}

	log.Debugf("AssociateRoute status code: [%v]", resp.StatusCode)
	return nil
}

func (c *CfAPI) UnassociateRoute(appID string, routeID string) error {
	address := fmt.Sprintf("%v/v2/apps/%v/routes/%v", c.BaseAddress, appID, routeID)
	err := c.deleteEntity(address, "route mapping")
	if err != nil {
		log.Errorf("Error unassociating route %v", routeID)
		return err
	}
	return nil
}

func (c *CfAPI) GetAppRoutes(appID string) (*types.CfRoutesResponse, error) {
	address := fmt.Sprintf("%v/v2/apps/%v/routes", c.BaseAddress, appID)
	response, err := c.getEntity(address, "routes")
	if err != nil {
		return nil, err
	}

	toReturn := new(types.CfRoutesResponse)
	json.NewDecoder(response.Body).Decode(toReturn)
	log.Debugf("Get routes status code: [%v]", response.StatusCode)
	return toReturn, nil
}

func (c *CfAPI) GetSpaceRoutesForHostname(spaceGUID, hostname string) (*types.CfRoutesResponse, error) {
	address := fmt.Sprintf("%v/v2/spaces/%v/routes?q=host:%v", c.BaseAddress, spaceGUID, hostname)
	response, err := c.getEntity(address, "routes")
	if err != nil {
		return nil, err
	}

	toReturn := new(types.CfRoutesResponse)
	json.NewDecoder(response.Body).Decode(toReturn)
	log.Debugf("Get routes status code: [%v]", response.StatusCode)
	log.Debugf("Retrieved %v route(s)", toReturn.Count)
	return toReturn, nil
}

func (c *CfAPI) GetAppsFromRoute(routeGUID string) (*types.CfAppsResponse, error) {
	address := fmt.Sprintf("%v/v2/routes/%v/apps", c.BaseAddress, routeGUID)
	response, err := c.getEntity(address, "apps")
	if err != nil {
		return nil, err
	}

	toReturn := new(types.CfAppsResponse)
	json.NewDecoder(response.Body).Decode(toReturn)
	log.Debugf("Get apps status code: [%v]", response.StatusCode)
	log.Debugf("Retrieved %v app(s)", toReturn.Count)
	return toReturn, nil
}

func (c *CfAPI) DeleteRoute(routeID string) error {
	address := fmt.Sprintf("%v/v2/routes/%v", c.BaseAddress, routeID)
	err := c.deleteEntity(address, "route")
	if err != nil {
		log.Errorf("Error deleting route %v", routeID)
		return err
	}
	return nil
}
