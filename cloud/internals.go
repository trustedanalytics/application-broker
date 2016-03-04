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
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/go-cf-lib/helpers"
	"github.com/trustedanalytics/go-cf-lib/types"
	"net/http"
	"strings"
	"sync"
)

// Clones user provided service with additional replacements of its content
func (cl *CloudAPI) CreateUserProvidedServiceClone(spaceGUID string, comp types.Component, suffix, url string,
	results chan types.ComponentClone, errorsCh chan error, wg *sync.WaitGroup) {

	defer wg.Done()
	serviceName := comp.Name + "-" + suffix
	log.Debugf("Create dependent user provided service: service=[%v])", serviceName)

	// Retrieve UPS
	response, err := cl.cf.GetUserProvidedService(comp.GUID)
	if err != nil {
		errorsCh <- err
		return
	}
	log.Infof("Dependent user provided service retrieved: %+v", response)

	// Create UPS
	response.Entity.Name = serviceName
	response.Entity.SpaceGUID = spaceGUID

	// Replace url to match clone application route
	if len(url) > 0 {
		response.Entity.Credentials["url"] = fmt.Sprintf("http://%v", url)
		response.Entity.Name = fmt.Sprintf("%v-ups", strings.Split(url, ".")[0])
	}
	// Generate random values where needed
	_ = cl.applyAdditionalReplacementsInUPSCredentials(response)

	response, err = cl.cf.CreateUserProvidedServiceInstance(&response.Entity)
	if err != nil {
		errorsCh <- err
		return
	}
	spawnedServiceInstanceGUID := response.Meta.GUID
	log.Debugf("Dependent user provided service created. Service Instance GUID=[%v]", spawnedServiceInstanceGUID)

	results <- types.ComponentClone{
		Component: comp,
		CloneGUID: spawnedServiceInstanceGUID,
	}
	errorsCh <- nil
	return
}

func (cl *CloudAPI) Discovery(sourceAppGUID string) ([]types.Component, error) {
	address := fmt.Sprintf("%v/v1/discover/%v", cl.appDepDiscUps.Url, sourceAppGUID)
	log.Infof("Getting application stack components: %v", address)

	client := &http.Client{}
	request, err := http.NewRequest("GET", address, nil)
	request.SetBasicAuth(cl.appDepDiscUps.AuthUser, cl.appDepDiscUps.AuthPass)
	response, err := client.Do(request)
	if err != nil {
		msg := fmt.Sprintf("Could not get application stack components: [%v]", err)
		log.Error(msg)
		return nil, types.InternalServerError{Context: msg}
	}

	if response.StatusCode == http.StatusNotFound {
		return nil, types.EntityNotFoundError{}
	}

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Get application stack components failed. Response from CC: (%d) [%v]",
			response.StatusCode, helpers.ReaderToString(response.Body))
		log.Error(msg)
		return nil, types.InternalServerError{Context: msg}
	}

	toReturn := make([]types.Component, 0)
	json.Unmarshal(helpers.ReaderToBytes(response.Body), &toReturn)
	log.Debugf("Get application stack components status code: [%v]", response.StatusCode)
	log.Debugf("Application stack components retrieved. Got %d results", len(toReturn))
	return toReturn, nil
}

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
	case nil:
		return true
	case types.EntityNotFoundError{}, types.InstanceNotFoundError{}, types.ServiceNotFoundError{}:
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
