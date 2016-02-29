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

package wrapper

import "github.com/trustedanalytics/go-cf-lib/types"

func (w *CfAPIWrapper) GetAppSummary(id string) (*types.CfAppSummary, error) {
	res, err := w.rest.GetAppSummary(id)
	return res, err
}

func (w *CfAPIWrapper) GetUserProvidedService(guid string) (*types.CfUserProvidedServiceResource, error) {
	res, err := w.rest.GetUserProvidedService(guid)
	return res, err
}

func (w *CfAPIWrapper) GetSpaceRoutesForHostname(spaceGUID, hostname string) (*types.CfRoutesResponse, error) {
	res, err := w.rest.GetSpaceRoutesForHostname(spaceGUID, hostname)
	return res, err
}

func (w *CfAPIWrapper) GetAppsFromRoute(routeGUID string) (*types.CfAppsResponse, error) {
	res, err := w.rest.GetAppsFromRoute(routeGUID)
	return res, err
}

func (w *CfAPIWrapper) GetServiceOfName(name string) (*types.CfServiceResource, error) {
	res, err := w.rest.GetServiceOfName(name)
	return res, err
}
