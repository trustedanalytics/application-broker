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
	"github.com/trustedanalytics/application-broker/misc/http-utils"
	"github.com/trustedanalytics/application-broker/types"
	"net/http"
	"time"
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

	Describe("create app method", func() {
		Context("with correct data passed", func() {
			It("should respond with guid", func() {
				httpmock.RegisterResponder("POST", "/v2/apps", positiveCreateAppResponder)

				sut := CfAPI{Client: http.DefaultClient}
				result, _ := sut.createApp(types.CfApp{Name: "appName"})

				Expect(result).NotTo(BeNil())
				Expect(result.Meta.GUID).To(Equal("super_fake_guid"))
			})
		})
	})

	Describe("associate app with route", func() {
		Context("with correct data passed", func() {
			It("should call cloudCtl and succeed", func() {

				appID := "fakeAppId"
				routeID := "fakeRouteId"
				address := fmt.Sprintf("/v2/apps/%v/routes/%v", appID, routeID)
				httpmock.RegisterResponder("PUT", address, positiveStatusCreatedResponder)

				sut := CfAPI{Client: http.DefaultClient}
				err := sut.associateRoute(appID, routeID)

				Expect(err).To(BeNil())
			})
		})
	})

	Describe("copy bits method", func() {

		destGUID, _ := uuid.NewV4()
		copyBitsAddress := fmt.Sprintf("/v2/apps/%v/copy_bits", destGUID)

		Context("when proper src dest guids given", func() {
			It("should copy bits on cloudCtl and wait for job success", func() {
				httpmock.RegisterResponder("POST", copyBitsAddress, positiveCopyBitsResponder)

				sut := CfAPI{Client: http.DefaultClient}
				asyncErr := make(chan error)
				go sut.copyBits("fake", destGUID.String(), asyncErr)

				Expect(<-asyncErr).ShouldNot(HaveOccurred())
			})
		})

		Context("when cloudCtl returns an error", func() {
			It("error is propagated", func() {
				httpmock.RegisterResponder("POST", copyBitsAddress, queuedCopyBitsResponder)
				httpmock.RegisterResponder("GET", "jobUrl", failedJobResponder)

				sut := CfAPI{Client: http.DefaultClient}
				asyncErr := make(chan error)
				go sut.copyBits("fake", destGUID.String(), asyncErr)

				err := <-asyncErr
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(misc.CcJobFailedError{"someErr"}))
			})
		})
	})

	Describe("restage app", func() {
		Context("in positive scenario", func() {
			It("should not return errors", func() {
				appGUID, _ := uuid.NewV4()
				restageAddress := fmt.Sprintf("/v2/apps/%v/restage", appGUID.String())
				httpmock.RegisterResponder("POST", restageAddress, positiveRestageResponder)

				sut := CfAPI{Client: http.DefaultClient}
				err := sut.restageApp(appGUID.String())

				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("update app", func() {
		executeTestCase := func(responder httpmock.Responder) error {
			guid, _ := uuid.NewV4()
			app := types.CfAppResource{}
			app.Meta.GUID = guid.String()
			updateURL := fmt.Sprintf("/v2/apps/%v", app.Meta.GUID)

			httpmock.RegisterResponder("PUT", updateURL, responder)
			sut := CfAPI{Client: http.DefaultClient}
			return sut.updateApp(&app)
		}

		Context("in positive scenario", func() {
			It("should not return error", func() {
				err := executeTestCase(positiveStatusCreatedResponder)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("in negative scenario", func() {
			It("should propagate error", func() {
				err := executeTestCase(negativeResponder)
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Describe("wait for app running", func() {
		Context("in positive scenario", func() {
			It("should not return error", func() {
				guid, _ := uuid.NewV4()
				app := types.CfAppResource{}
				app.Meta.GUID = guid.String()
				instances := fmt.Sprintf("/v2/apps/%v/instances", app.Meta.GUID)
				httpmock.RegisterResponder("GET", instances, positiveGetInstancesResponder)

				asyncErr := make(chan error)
				sut := CfAPI{Client: http.DefaultClient}
				go sut.waitForAppRunning(app.Meta.GUID, asyncErr)

				select {
				case err := <-asyncErr:
					Expect(err).ShouldNot(HaveOccurred())
				case <-time.After(5 * time.Millisecond):
					Fail("waitAppForRunning entered infinite loop")
				}
			})
		})
	})

	Describe("service deprovision", func() {
		var (
			sut            CfAPI
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
			httpmock.RegisterResponder(httputils.MethodDelete, appURL, responderGenerator(204, nil))
			appSummaryURL = fmt.Sprintf("/v2/apps/%v/summary", appGUID)
			httpmock.RegisterResponder(httputils.MethodGet, appSummaryURL, responderGenerator(200, app))

			sut = CfAPI{Client: http.DefaultClient}
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

		Context("Get app summary fails", func() {
			It("should forward error", func() {
				httpmock.RegisterResponder("GET", appSummaryURL, responderGenerator(500, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Get application summary failed"))
			})
		})

		Context("CF fails on app deletion", func() {
			It("should forward error", func() {
				httpmock.RegisterResponder("DELETE", appURL, responderGenerator(500, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Delete application failed"))
			})
		})

		Context("Get app bindings fails", func() {
			It("should forward error", func() {
				httpmock.RegisterResponder("GET", appBindingsURL, responderGenerator(500, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Get app bindings failed"))
			})
		})

		Context("Get app bindings fails", func() {
			It("should forward error", func() {
				httpmock.RegisterResponder("GET", appBindingsURL, responderGenerator(500, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Get app bindings failed"))
			})
		})

		Context("Binding does not exist", func() {
			It("should continue silently", func() {
				registerServiceUnbind(appGUID, bindings.Resources[1].Meta.GUID, responderGenerator(404, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Delete of a binding fails", func() {
			It("should forward error", func() {
				registerServiceUnbind(appGUID, bindings.Resources[1].Meta.GUID, responderGenerator(500, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Delete binding failed"))
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
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Delete route mapping failed"))
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
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Delete route failed"))
			})
		})

		Context("Service does not exist", func() {
			It("Should continue silently", func() {
				registerServiceDelete(bindings.Resources[1].Entity.ServiceInstanceGUID, responderGenerator(404, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Delete of a service instance fails", func() {
			It("should forward error", func() {
				registerServiceDelete(bindings.Resources[1].Entity.ServiceInstanceGUID, responderGenerator(500, nil))

				err := sut.Deprovision(appGUID)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Delete service instance failed"))
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
	httpmock.RegisterResponder(httputils.MethodDelete, url, response)
}

func registerServiceDelete(instanceGUID string, response httpmock.Responder) {
	url := fmt.Sprintf("/v2/service_instances/%v", instanceGUID)
	httpmock.RegisterResponder(httputils.MethodDelete, url, response)
}

func registerRouteUnbind(appGUID string, routeGUID string, response httpmock.Responder) {
	url := fmt.Sprintf("/v2/apps/%v/routes/%v", appGUID, routeGUID)
	httpmock.RegisterResponder(httputils.MethodDelete, url, response)
}

func registerRouteDelete(routeGUID string, response httpmock.Responder) {
	url := fmt.Sprintf("/v2/routes/%v", routeGUID)
	httpmock.RegisterResponder(httputils.MethodDelete, url, response)
}

func newRoute() types.CfAppSummaryRoute {
	return types.CfAppSummaryRoute{GUID: misc.NewGUID()}
}
