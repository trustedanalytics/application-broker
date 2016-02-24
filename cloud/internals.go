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
	"encoding/json"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
)

func (cl *CloudAPI) groupComponentsByType(order []types.Component) map[types.ComponentType][]types.Component {
	groupedComponents := make(map[types.ComponentType][]types.Component)
	groupedComponents[types.ComponentApp] = []types.Component{}
	groupedComponents[types.ComponentUPS] = []types.Component{}
	groupedComponents[types.ComponentService] = []types.Component{}
	for _, comp := range order {
		if _, ok := groupedComponents[comp.Type]; ok {
			groupedComponents[comp.Type] = append(groupedComponents[comp.Type], comp)
		} else {
			groupedComponents[comp.Type] = []types.Component{comp}
		}
	}
	log.Infof("%+v", groupedComponents)
	return groupedComponents
}

func (cl *CloudAPI) isErrorAcceptedDuringDeprovision(err error) bool {
	switch err {
	case misc.EntityNotFoundError{}, misc.InstanceNotFoundError{}, misc.ServiceNotFoundError{}:
		log.Errorf("Accepted error occured during deprovisioning: %v", err.Error())
		return true
	}
	return false
}

func (cl *CloudAPI) applyAdditionalReplacementsInUPSCredentials(response *types.CfUserProvidedServiceResource) error {
	credentials, err := json.Marshal(response.Entity.Credentials)
	if err != nil {
		return err
	}
	credentialsStr := string(credentials)
	credentialsStr = misc.ReplaceWithRandom(credentialsStr)
	log.Infof("Final UPS %v content %v", response.Entity.Name, credentialsStr)
	json.Unmarshal([]byte(credentialsStr), &response.Entity.Credentials)
	return nil
}
