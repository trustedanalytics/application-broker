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

package broker

import (
	"bytes"
	"encoding/json"
	"github.com/cloudfoundry-community/types-cf"
	"github.com/go-martini/martini"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"github.com/trustedanalytics/application-broker/dao"
	"github.com/trustedanalytics/application-broker/messagebus"
	"github.com/trustedanalytics/application-broker/service"
	"github.com/trustedanalytics/application-broker/service/extension"
	"github.com/trustedanalytics/go-cf-lib/types"
	"net/http"
	"os"
	"strings"
)

var _ = Describe("Handler", func() {

	var (
		sut       *handler
		mongoMock *dao.FacadeMock
		cfMock    *service.CfMock
	)

	BeforeEach(func() {
		mongoMock = new(dao.FacadeMock)
		cfMock = new(service.CfMock)
		cfMock.On("UpdateBroker", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		os.Setenv("VCAP_APPLICATION", "{\"name\":\"banana\",\"uris\":[\"http://fakeurl\"]}")
	})

	JustBeforeEach(func() {
		svc := service.New(
			mongoMock,
			cfMock,
			new(messagebus.DevNullBus),
			service.CreationStatusFactory{},
		)
		sut = newHandler(svc)
	})

	AfterEach(func() {
		os.Unsetenv("VCAP_APPLICATION")
	})

	Describe("when appending new service to catalog", func() {
		Context("with unmarshallable body", func() {
			It("should return bad request", func() {
				unmarshallableBody := bytes.NewReader([]byte("123"))
				req, _ := http.NewRequest("", "", unmarshallableBody)

				code, _ := sut.append(req, nil)

				Expect(code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("with correct service data passed", func() {

			BeforeEach(func() {
				mongoMock.On("Append", mock.Anything).Return()
				cfMock.On("CheckIfServiceExists", "dummy").Return(nil)
			})

			It("should return status created", func() {
				jsonString := `{"name":"dummy", "description":"dummier", "app":{"metadata" : {"guid":"fake"}}}`
				correctBody := strings.NewReader(jsonString)
				req, _ := http.NewRequest("", "", correctBody)

				code, resp := sut.append(req, nil)

				decoded := extension.ServiceExtension{}
				json.NewDecoder(strings.NewReader(resp)).Decode(&decoded)
				Expect(decoded.ID).NotTo(BeEmpty())
				Expect(decoded.Plans).NotTo(BeNil())
				Expect(code).To(Equal(http.StatusCreated))
			})
		})
	})

	Describe("deleting new service from catalog", func() {
		Context("when service does not exist", func() {
			It("404 not found should be returned", func() {
				params := martini.Params{"service_id": "not-existing-id"}
				mongoMock.On("Find", "not-existing-id").Return(nil, types.ServiceNotFoundError)

				code, _ := sut.remove(nil, params)
				Expect(code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("deleting new service from catalog", func() {
		Context("in positive scenario", func() {
			It("204 no content should be returned", func() {
				params := martini.Params{"service_id": "existing-id"}
				mongoMock.On("Find", "existing-id").Return(&extension.ServiceExtension{}, nil)
				mongoMock.On("HasInstancesOf", "existing-id").Return(false, nil)
				mongoMock.On("Remove", "existing-id").Return(nil)
				mongoMock.On("Get").Return([]*extension.ServiceExtension{new(extension.ServiceExtension), new(extension.ServiceExtension)})

				code, _ := sut.remove(nil, params)
				Expect(code).To(Equal(http.StatusNoContent))
			})
		})
	})

	Describe("when retrieving services from catalog", func() {
		var (
			fakeCatalog []*extension.ServiceExtension
			resp        extension.CatalogExtension
		)

		Context("and catalog contais services", func() {

			BeforeEach(func() {
				fakeCatalog = append(fakeCatalog, &extension.ServiceExtension{})
				fakeCatalog = append(fakeCatalog, &extension.ServiceExtension{})
				mongoMock.On("Get").Return(fakeCatalog)
			})

			It("should return list of services", func() {
				req, _ := http.NewRequest("", "", nil)
				code, raw := sut.catalog(req, nil)
				json.NewDecoder(strings.NewReader(raw)).Decode(&resp)

				Expect(len(resp.Services)).NotTo(Equal(0))
				Expect(code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("when provisioning new service instance", func() {
		var (
			resp        extension.ServiceCreationResponse
			testService extension.ServiceExtension
		)

		BeforeEach(func() {
			inner := cf.Service{ID: "fakeId"}
			testService = extension.ServiceExtension{
				Service:      inner,
				ReferenceApp: types.CfAppResource{Entity: types.CfApp{Name: "appToClone"}},
			}
			mongoMock.On("Find", inner.ID).Return(&testService)
			mongoMock.On("AppendInstance", mock.Anything).Return()
			cfMock.On("Provision", testService.ReferenceApp.Meta.GUID, mock.Anything, mock.Anything).Return(&extension.ServiceCreationResponse{})
		})

		Context("and requested service type exists", func() {
			It("should return service creation response", func() {
				bytesToRead, _ := json.Marshal(cf.ServiceCreationRequest{ServiceID: testService.Service.ID})
				correctBody := bytes.NewReader(bytesToRead)

				req, _ := http.NewRequest("", "", correctBody)
				code, raw := sut.provision(req, nil)
				json.NewDecoder(strings.NewReader(raw)).Decode(&resp)

				Expect(resp).NotTo(BeNil())
				Expect(code).To(Equal(http.StatusCreated))
			})
		})
	})

	Describe("when binding service instance", func() {

		BeforeEach(func() {
			svcInstance := extension.ServiceInstanceExtension{
				App: types.CfAppResource{Meta: types.CfMeta{URL: "someUrl"}},
			}
			mongoMock.On("FindInstance", "fakeInstanceID").Return(&svcInstance)
		})

		Context("with app to bind to", func() {
			It("should return StatusCreated", func() {
				request := cf.ServiceBindingRequest{
					AppGUID: "fakeAppToBindTo",
				}
				bytesToRead, _ := json.Marshal(request)
				correctBody := bytes.NewReader(bytesToRead)
				req, _ := http.NewRequest("", "", correctBody)
				code, _ := sut.bind(req, martini.Params{"instance_id": "fakeInstanceID"})

				Expect(code).To(Equal(http.StatusCreated))
			})

			It("should return url in credentials", func() {
				request := cf.ServiceBindingRequest{
					AppGUID: "fakeAppToBindTo",
				}
				bytesToRead, _ := json.Marshal(request)
				correctBody := bytes.NewReader(bytesToRead)
				req, _ := http.NewRequest("", "", correctBody)
				_, raw := sut.bind(req, martini.Params{"instance_id": "fakeInstanceID"})

				resp := types.ServiceBindingResponse{}
				json.NewDecoder(strings.NewReader(raw)).Decode(&resp)
				Expect(resp.Credentials["url"]).To(Equal("someUrl"))
			})
		})

		Context("without app to bind to", func() {
			It("should return Status OK", func() {
				request := cf.ServiceBindingRequest{}
				bytesToRead, _ := json.Marshal(request)
				correctBody := bytes.NewReader(bytesToRead)
				req, _ := http.NewRequest("", "", correctBody)
				code, _ := sut.bind(req, martini.Params{"instance_id": "fakeInstanceID"})

				Expect(code).To(Equal(http.StatusOK))
			})
		})
	})
})
