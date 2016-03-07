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
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/go-cf-lib/helpers"
	"github.com/trustedanalytics/go-cf-lib/types"
	"sync"
)

func (w *CfAPI) DeleteServiceInstIfUnbound(comp types.Component,
	errorsCh chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()

	bindings, err := w.GetServiceBindings(comp.GUID)
	if err != nil {
		errorsCh <- err
		return
	}
	if bindings.TotalResults == 0 {
		log.Infof("Service %v is not bound to anything", comp.Name)
		log.Infof("Deleting %v instance %v", comp.Type, comp.Name)
		if err := w.DeleteServiceInstance(comp.GUID); err != nil {
			errorsCh <- err
			return
		}
	} else {
		log.Infof("%v instance %v is bound to %v apps. Not deleting instance.",
			comp.Type, comp.Name, bindings.TotalResults)
	}
	errorsCh <- nil
}

func (w *CfAPI) DeleteUPSInstIfUnbound(comp types.Component,
	errorsCh chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()

	bindings, err := w.GetUserProvidedServiceBindings(comp.GUID)
	if err != nil {
		errorsCh <- err
		return
	}
	if bindings.TotalResults == 0 {
		log.Infof("Service %v is not bound to anything", comp.Name)
		log.Infof("Deleting %v instance %v", comp.Type, comp.Name)
		if err := w.DeleteUserProvidedServiceInstance(comp.GUID); err != nil {
			errorsCh <- err
			return
		}
	} else {
		log.Infof("%v instance %v is bound to %v apps. Not deleting instance.",
			comp.Type, comp.Name, bindings.TotalResults)
	}
	errorsCh <- nil
}

func (w *CfAPI) DeleteRoutes(appGUID string, errorsCh chan error, doneWaitGroup *sync.WaitGroup) {
	defer doneWaitGroup.Done()

	appSummary, _ := w.GetAppSummary(appGUID)
	if appSummary == nil {
		// Application not exist so no routes to remove
		log.Infof("Application already does not exist so no routes should be deleted")
		errorsCh <- nil
		return
	}
	routes := appSummary.Routes

	wg := sync.WaitGroup{}
	wg.Add(len(routes))
	results := make(chan error, len(routes))

	for _, loopRoute := range routes {
		go func(route types.CfAppSummaryRoute) {
			defer wg.Done()
			if err := w.UnassociateRoute(appGUID, route.GUID); err != nil {
				results <- err
				return
			}
			if err := w.DeleteRoute(route.GUID); err != nil {
				results <- err
				return
			}
			results <- nil
		}(loopRoute)
	}
	wg.Wait()
	errorsCh <- helpers.FirstNonEmpty(results, len(routes))
}
