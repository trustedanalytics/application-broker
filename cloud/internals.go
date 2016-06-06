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
	"github.com/signalfx/golib/errors"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/service/extension"
	"github.com/trustedanalytics/go-cf-lib/helpers"
	"github.com/trustedanalytics/go-cf-lib/types"
	"net/http"
	"strings"
	"sync"
)

// Clones user provided service with additional replacements of its content
func (cloud *CloudAPI) CreateUserProvidedServiceClone(spaceGUID string, comp types.Component, suffix, url string,
	results chan types.ComponentClone, errorsCh chan error, wg *sync.WaitGroup) {

	defer wg.Done()
	serviceName := comp.Name + "-" + suffix
	log.Debugf("Create dependent user provided service: service=[%v])", serviceName)

	// Retrieve UPS
	response, err := cloud.cf.GetUserProvidedService(comp.GUID)
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
	_ = cloud.applyAdditionalReplacementsInUPSCredentials(response)

	response, err = cloud.cf.CreateUserProvidedServiceInstance(&response.Entity)
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

func (cloud *CloudAPI) Discovery(sourceAppGUID string) ([]types.Component, error) {
	address := fmt.Sprintf("%v/v1/discover/%v", cloud.appDepDiscUps.Url, sourceAppGUID)
	log.Infof("Getting application stack components: %v", address)

	client := &http.Client{}
	request, err := http.NewRequest("GET", address, nil)
	request.SetBasicAuth(cloud.appDepDiscUps.AuthUser, cloud.appDepDiscUps.AuthPass)
	response, err := client.Do(request)
	if err != nil {
		msg := fmt.Sprintf("Could not get application stack components: [%v]", err)
		log.Error(msg)
		return nil, errors.Annotate(types.InternalServerError, msg)
	}

	if response.StatusCode == http.StatusNotFound {
		return nil, types.EntityNotFoundError
	}

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Get application stack components failed. Response from CC: (%d) [%v]",
			response.StatusCode, helpers.ReaderToString(response.Body))
		log.Error(msg)
		return nil, errors.Annotate(types.InternalServerError, msg)
	}

	toReturn := make([]types.Component, 0)
	json.Unmarshal(helpers.ReaderToBytes(response.Body), &toReturn)
	log.Debugf("Get application stack components status code: [%v]", response.StatusCode)
	log.Debugf("Application stack components retrieved. Got %d results", len(toReturn))
	return toReturn, nil
}

func (cloud *CloudAPI) groupComponentsByType(order []types.Component) map[types.ComponentType][]types.Component {
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

func (cloud *CloudAPI) selectAcceptedServiceParams(serviceName string,
	passedParams map[string]string,
	allServicesConfigurations []*extension.ServiceConfiguration) map[string]interface{} {

	if passedParams == nil || allServicesConfigurations == nil {
		log.Info("No additional service parameters passed or services not configurable")
		return nil
	}

	serviceConf := cloud.selectServiceConf(serviceName, allServicesConfigurations)
	if serviceConf == nil {
		log.Infof("Service %v does not accept additional configuration", serviceName)
		return nil
	}

	paramsToPass := make(map[string]interface{})
	for k, v := range passedParams {
		parts := strings.SplitN(k, ".", 2)
		if len(parts) != 2 {
			// without namespace
			cloud.addParamIfConfigurable(k, v, serviceConf, paramsToPass)
			continue
		}
		namespace := parts[0]
		key := parts[1]
		if namespace != serviceName {
			// param not for this service
			continue
		}
		// configurable param
		cloud.addParamIfConfigurable(key, v, serviceConf, paramsToPass)
	}
	if len(paramsToPass) == 0 {
		return nil
	}
	return paramsToPass
}

func (cloud *CloudAPI) selectServiceConf(serviceName string,
	allServicesConfigurations []*extension.ServiceConfiguration) *extension.ServiceConfiguration {

	for _, conf := range allServicesConfigurations {
		if conf.ServiceName == serviceName {
			return conf
		}
	}
	return nil
}

func (cloud *CloudAPI) addParamIfConfigurable(key, value string,
	serviceConf *extension.ServiceConfiguration,
	paramsToPass map[string]interface{}) {

	for _, allowedParam := range serviceConf.Params {
		if key == allowedParam {
			paramsToPass[allowedParam] = value
			break
		}
	}
}

func (cloud *CloudAPI) removeParametersNamespaces(passed map[string]string) (map[string]string, error) {
	if passed == nil {
		return nil, nil
	}
	noNamespace := make(map[string]string)
	possibleCollisions := make(map[string]bool)
	for k, v := range passed {
		a := strings.SplitN(k, ".", 2)
		if len(a) != 2 {
			if _, ok := noNamespace[k]; !ok {
				noNamespace[k] = v
				possibleCollisions[k] = true
			} else {
				return nil, errors.Errorf("Colision of keys in additional parameters provided. Please use namespaces for key %v", k)
			}
		} else {
			if _, ok := noNamespace[a[1]]; !ok {
				noNamespace[a[1]] = v
			} else if _, ok := possibleCollisions[a[1]]; ok {
				return nil, errors.Errorf("Colision of keys in additional parameters provided. Please use namespaces for key %v", a[1])
			}
		}

	}
	return noNamespace, nil
}

func (cloud *CloudAPI) isErrorAcceptedDuringDeprovision(err error) bool {
	switch err {
	case nil:
		return true
	case types.EntityNotFoundError, types.InstanceNotFoundError, types.ServiceNotFoundError:
		log.Errorf("Accepted error occured during deprovisioning: %v", err.Error())
		return true
	}
	return false
}

func (cloud *CloudAPI) applyAdditionalReplacementsInUPSCredentials(response *types.CfUserProvidedServiceResource) error {
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

func (cloud *CloudAPI) logParameters(parameters map[string]string, servicesConfiguration []*extension.ServiceConfiguration) {
	log.Infof("Additional parameters passed: %v", parameters)
	if len(servicesConfiguration) > 0 {
		log.Infof("Configurable service parameters:")
		for _, conf := range servicesConfiguration {
			log.Infof("%v", *conf)
		}
	} else {
		log.Info("No configurable service parameters")
	}
}
