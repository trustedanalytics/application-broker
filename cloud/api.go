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
	"github.com/cloudfoundry-community/types-cf"
	"github.com/trustedanalytics/application-broker/types"
)

type API interface {
	Provision(sourceAppGUID string, request *cf.ServiceCreationRequest) (*types.ServiceCreationResponse, error)
	Deprovision(appGUID string) error
	UpdateBroker(brokerName string, brokerURL string, username string, password string) error
}
