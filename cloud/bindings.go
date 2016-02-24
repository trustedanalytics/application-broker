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
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"sync"
)

func (cl *CloudAPI) bindService(appGUID, serviceGUID string, wg *sync.WaitGroup, errors chan error) {
	defer wg.Done()
	// Bind created service
	svcBindingReq := types.NewCfServiceBindingRequest(appGUID, serviceGUID)
	svcBindingResp, err := cl.cf.CreateServiceBinding(svcBindingReq)
	if err != nil {
		errors <- err
		return
	}
	log.Debugf("Dependent service binding created: Service Binding GUID=[%v]", svcBindingResp.Meta.GUID)
	errors <- nil
	return
}

func (cl *CloudAPI) unbindAppServices(appGUID string, result chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()

	bindings, err := cl.cf.GetAppBindigs(appGUID)
	if err != nil {
		result <- err
		return
	}
	var results = make(chan error, len(bindings.Resources))
	wg := sync.WaitGroup{}
	wg.Add(len(bindings.Resources))

	for _, loopBinding := range bindings.Resources {
		go func(binding types.CfBindingResource) {
			defer wg.Done()
			if err := cl.cf.DeleteBinding(binding); err != nil {
				results <- err
				return
			}
			results <- nil
		}(loopBinding)
	}
	wg.Wait()
	result <- misc.FirstNonEmpty(results, len(bindings.Resources))
}
