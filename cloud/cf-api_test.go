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
	"fmt"
	"github.com/jarcoal/httpmock"
	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/go-cf-lib/api"
	"github.com/trustedanalytics/go-cf-lib/types"
	"net/http"
)

var _ = Describe("Cf api", func() {

	var (
		positiveCreateAppResponder     httpmock.Responder
		positiveStatusCreatedResponder httpmock.Responder
		positiveCopyBitsResponder      httpmock.Responder
		positiveRestageResponder       httpmock.Responder
		positiveGetInstancesResponder  httpmock.Responder
		queuedCopyBitsResponder        httpmock.Responder
		negativeResponder              httpmock.Responder
		failedJobResponder             httpmock.Responder
	)

	BeforeEach(func() {
		httpmock.Activate()

		appToReturn := types.CfAppResource{}
		appToReturn.Meta = types.CfMeta{GUID: "super_fake_guid"}
		positiveCreateAppResponder = responderGenerator(201, appToReturn)

		positiveStatusCreatedResponder = responderGenerator(201, nil)

		finishedJob := types.CfJobResponse{Entity: types.CfJob{Status: "finished"}}
		positiveCopyBitsResponder = responderGenerator(201, finishedJob)

		queuedJob := types.CfJobResponse{
			Entity: types.CfJob{Status: "queued"},
			Meta:   types.CfMeta{URL: "jobUrl"},
		}
		queuedCopyBitsResponder = responderGenerator(201, queuedJob)

		failJob := types.CfJobResponse{
			Entity: types.CfJob{Status: "failed", Error: "someErr"},
		}
		failedJobResponder = responderGenerator(200, failJob)

		restagedApp := types.CfAppResource{Entity: types.CfApp{State: "STARTED"}}
		positiveRestageResponder = responderGenerator(201, restagedApp)

		instances := map[string]types.CfAppInstance{}
		instances["0"] = types.CfAppInstance{State: "RUNNING"}
		positiveGetInstancesResponder = responderGenerator(200, instances)

		negativeResponder = responderGenerator(400, nil)
	})

	AfterEach(func() {
		httpmock.DeactivateAndReset()
	})

	Describe("service deprovision", func() {
		var (
			sut            *CloudAPI
			app            types.CfAppSummary
			appGUID        string
			appURL         string
			appSummaryURL  string
			appBindingsURL string
			bindings       types.CfBindingsResources
		)

		BeforeEach(func() {
			httpmock.Activate()

			guid, _ := uuid.NewV4()
			appGUID = guid.String()

			app = types.CfAppSummary{}
			app.GUID = appGUID

			bindings, appBindingsURL = registerServiceCleanupRequests(appGUID)
			app.Routes = registerRoutesCleanupRequests(appGUID)

			appURL = fmt.Sprintf("/v2/apps/%v", appGUID)
			httpmock.RegisterResponder(api.MethodDelete, appURL, responderGenerator(204, nil))
			appSummaryURL = fmt.Sprintf("/v2/apps/%v/summary", appGUID)
			httpmock.RegisterResponder(api.MethodGet, appSummaryURL, responderGenerator(200, app))

			sut = NewCloudAPI(nil)
		})

		AfterEach(func() {
			httpmock.DeactivateAndReset()
		})

		Context("app doesn't exist", func() {
			It("should process as normal", func() {
				httpmock.RegisterResponder("GET", appSummaryURL, responderGenerator(404, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Binding does not exist", func() {
			It("should continue silently", func() {
				registerServiceUnbind(appGUID, bindings.Resources[1].Meta.GUID, responderGenerator(404, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Route mapping does not exist", func() {
			It("should continue silently", func() {
				registerRouteUnbind(appGUID, app.Routes[1].GUID, responderGenerator(404, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Delete of a route mapping fails", func() {
			It("should forward error", func() {
				registerRouteUnbind(appGUID, app.Routes[1].GUID, responderGenerator(500, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Route does not exist", func() {
			It("should continue silently", func() {
				registerRouteDelete(app.Routes[1].GUID, responderGenerator(404, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Delete of a route fails", func() {
			It("should forward error", func() {
				registerRouteDelete(app.Routes[1].GUID, responderGenerator(500, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Service does not exist", func() {
			It("Should continue silently", func() {
				registerServiceDelete(bindings.Resources[1].Entity.ServiceInstanceGUID, responderGenerator(404, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Everything ok", func() {
			It("should return OK", func() {
				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

})

func responderGenerator(code int, v interface{}) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		resp, _ := httpmock.NewJsonResponse(code, v)
		return resp, nil
	}
}

func newBinding(appGUID string) types.CfBindingResource {
	return types.CfBindingResource{
		Meta:   types.CfMeta{GUID: misc.NewGUID()},
		Entity: types.CfBinding{AppGUID: appGUID, ServiceInstanceGUID: misc.NewGUID()}}
}

func registerRoutesCleanupRequests(appGUID string) []types.CfAppSummaryRoute {
	routes := []types.CfAppSummaryRoute{newRoute(), newRoute()}

	for _, route := range routes {
		registerRouteUnbind(appGUID, route.GUID, responderGenerator(204, nil))
		registerRouteDelete(route.GUID, responderGenerator(204, nil))
	}
	return routes
}

func registerServiceCleanupRequests(appGUID string) (types.CfBindingsResources, string) {
	bindings := types.CfBindingsResources{
		Resources:    []types.CfBindingResource{newBinding(appGUID), newBinding(appGUID), newBinding(appGUID)},
		TotalResults: 3}
	appBindingsURL := fmt.Sprintf("/v2/apps/%v/service_bindings", appGUID)
	httpmock.RegisterResponder("GET", appBindingsURL, responderGenerator(200, bindings))

	for _, binding := range bindings.Resources {
		registerServiceUnbind(appGUID, binding.Meta.GUID, responderGenerator(204, nil))
		registerServiceDelete(binding.Entity.ServiceInstanceGUID, responderGenerator(204, nil))
	}
	return bindings, appBindingsURL
}

func registerServiceUnbind(appGUID string, bindingGUID string, response httpmock.Responder) {
	url := fmt.Sprintf("/v2/apps/%v/service_bindings/%v", appGUID, bindingGUID)
	httpmock.RegisterResponder(api.MethodDelete, url, response)
}

func registerServiceDelete(instanceGUID string, response httpmock.Responder) {
	url := fmt.Sprintf("/v2/service_instances/%v", instanceGUID)
	httpmock.RegisterResponder(api.MethodDelete, url, response)
}

func registerRouteUnbind(appGUID string, routeGUID string, response httpmock.Responder) {
	url := fmt.Sprintf("/v2/apps/%v/routes/%v", appGUID, routeGUID)
	httpmock.RegisterResponder(api.MethodDelete, url, response)
}

func registerRouteDelete(routeGUID string, response httpmock.Responder) {
	url := fmt.Sprintf("/v2/routes/%v", routeGUID)
	httpmock.RegisterResponder(api.MethodDelete, url, response)
}

func newRoute() types.CfAppSummaryRoute {
	return types.CfAppSummaryRoute{GUID: misc.NewGUID()}
}
