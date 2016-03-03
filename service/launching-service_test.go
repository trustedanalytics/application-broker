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

package service

import (
	"errors"
	"github.com/cloudfoundry-community/types-cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"github.com/trustedanalytics/application-broker/dao"
	"github.com/trustedanalytics/application-broker/messagebus"
	"github.com/trustedanalytics/application-broker/service/extension"
	"github.com/trustedanalytics/go-cf-lib/types"
	"os"
)

var _ = Describe("Launching service", func() {

	var (
		dataCatalog *dao.FacadeMock
		nats        messagebus.MessageBus
		cfMock      *CfMock
	)

	BeforeEach(func() {
		dataCatalog = new(dao.FacadeMock)
		nats = new(messagebus.DevNullBus)
		os.Setenv("VCAP_APPLICATION", "{\"name\":\"banana\",\"uris\":[\"http://fakeurl\"]}")
		cfMock = new(CfMock)
		cfMock.On("UpdateBroker", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	})

	AfterEach(func() {
		os.Unsetenv("VCAP_APPLICATION")
	})

	Describe("get catalog", func() {
		Context("when catalog is empty", func() {
			It("should return zero services", func() {
				dataCatalog.On("Get").Return([]*extension.ServiceExtension{})

				sut := New(dataCatalog, nil, nats, CreationStatusFactory{})
				catalog, _ := sut.GetCatalog()

				dataCatalog.AssertNumberOfCalls(GinkgoT(), "Get", 1)
				Expect(len(catalog.Services)).To(Equal(0))
			})
		})

		Context("when catalog filled with one service", func() {
			It("should return one service", func() {
				services := []*extension.ServiceExtension{}
				services = append(services, &extension.ServiceExtension{
					ReferenceApp: types.CfAppResource{Meta: types.CfMeta{GUID: "someId"}},
				})
				dataCatalog.On("Get", mock.Anything).Return(services, nil)

				sut := New(dataCatalog, nil, nats, CreationStatusFactory{})
				catalog, _ := sut.GetCatalog()

				dataCatalog.AssertNumberOfCalls(GinkgoT(), "Get", 1)
				Expect(len(catalog.Services)).To(Equal(1))
			})
		})
	})

	Describe("append", func() {
		Context("not valid service", func() {
			It("should return error indicating bad input", func() {
				dataCatalog.On("Append", mock.Anything).Return()
				service := &extension.ServiceExtension{
					ReferenceApp: types.CfAppResource{Meta: types.CfMeta{GUID: "someId"}},
				}
				sut := New(dataCatalog, nil, nats, CreationStatusFactory{})
				err := sut.InsertToCatalog(service)

				Expect(err).To(Equal(types.InvalidInputError{}))
				dataCatalog.AssertNumberOfCalls(GinkgoT(), "Append", 0)
			})
		})

		Context("valid service", func() {
			It("should return non empty response", func() {
				basic := cf.Service{Name: "someName", Description: "desc"}
				service := &extension.ServiceExtension{
					ReferenceApp: types.CfAppResource{Meta: types.CfMeta{GUID: "someId"}},
					Service:      basic,
				}
				dataCatalog.On("Append", service).Return()
				cfMock.On("CheckIfServiceExists", service.Name).Return(nil)

				sut := New(dataCatalog, cfMock, nats, CreationStatusFactory{})
				err := sut.InsertToCatalog(service)

				Expect(err).To(BeNil())
				dataCatalog.AssertNumberOfCalls(GinkgoT(), "Append", 1)
			})
		})
	})

	Describe("delete from catalog", func() {
		Context("not existing service", func() {
			It("should return error", func() {
				expectedError := errors.New("No such service")
				cfApi := new(CfMock)
				dataCatalog.On("Find", "fakeId").Return(nil, expectedError)
				sut := New(dataCatalog, cfApi, nats, CreationStatusFactory{})

				err := sut.DeleteFromCatalog("fakeId")
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(expectedError))
			})
		})

		Context("service that has instances", func() {
			It("should return error", func() {
				fakeService := extension.ServiceExtension{Service: cf.Service{ID: "ID"}}
				dataCatalog.On("Find", "fakeId").Return(&fakeService, nil)
				dataCatalog.On("HasInstancesOf", "fakeId").Return(true, nil)
				dataCatalog.On("Get").Return([]*extension.ServiceExtension{new(extension.ServiceExtension), new(extension.ServiceExtension)})

				sut := New(dataCatalog, nil, nats, CreationStatusFactory{})

				err := sut.DeleteFromCatalog("fakeId")
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(types.ExistingInstancesError{}))
			})
		})

		Context("service without instances", func() {
			It("should succeed", func() {
				fakeService := extension.ServiceExtension{Service: cf.Service{ID: "ID"}}
				dataCatalog.On("Find", "fakeId").Return(&fakeService, nil)
				dataCatalog.On("HasInstancesOf", "fakeId").Return(false, nil)
				dataCatalog.On("Remove", "fakeId").Return(nil)
				dataCatalog.On("Get").Return([]*extension.ServiceExtension{new(extension.ServiceExtension), new(extension.ServiceExtension)})

				sut := New(dataCatalog, cfMock, nats, CreationStatusFactory{})

				err := sut.DeleteFromCatalog("fakeId")
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("the only service in catalog", func() {
			It("should throw an error", func() {
				fakeService := extension.ServiceExtension{Service: cf.Service{ID: "ID"}}
				dataCatalog.On("Find", "fakeId").Return(&fakeService, nil)
				dataCatalog.On("HasInstancesOf", "fakeId").Return(false, nil)
				dataCatalog.On("Remove", "fakeId").Return(nil)
				dataCatalog.On("Get").Return([]*extension.ServiceExtension{new(extension.ServiceExtension)})

				sut := New(dataCatalog, nil, nats, CreationStatusFactory{})

				err := sut.DeleteFromCatalog("fakeId")
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Describe("create service", func() {
		Context("in case of cloud foundry error", func() {
			//TODO:make this test simplier, shorter, etc...
			It("should propagate error", func() {
				svc := cf.Service{Name: "super_service"}
				svcExt := &extension.ServiceExtension{
					ReferenceApp: types.CfAppResource{Meta: types.CfMeta{GUID: "source_app_id"}},
					Service:      svc,
				}
				dataCatalog.On("Find", mock.Anything).Return(svcExt)
				request := new(cf.ServiceCreationRequest)
				request.Parameters = make(map[string]string)

				cfApi := new(CfMock)
				expectedErr := errors.New("ERROR!")
				cfApi.On("Provision", mock.Anything, mock.Anything).Return(nil, expectedErr)

				sut := New(dataCatalog, cfApi, nats, CreationStatusFactory{})
				resp, err := sut.CreateService(request)

				cfApi.AssertExpectations(GinkgoT())
				Expect(resp).To(BeNil())
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when provisioning succeeds", func() {
			//TODO:make this test simplier, shorter, etc...
			It("should return non empty response", func() {
				svc := cf.Service{Name: "super_service"}
				svcExt := &extension.ServiceExtension{
					ReferenceApp: types.CfAppResource{Meta: types.CfMeta{GUID: "source_app_id"}},
					Service:      svc,
				}
				dataCatalog.On("Find", "service_id").Return(svcExt)
				dataCatalog.On("AppendInstance", mock.Anything).Return()
				request := new(cf.ServiceCreationRequest)
				request.SpaceGUID = "space_guid"
				request.ServiceID = "service_id"
				request.Parameters = make(map[string]string)

				cfApi := new(CfMock)
				createAppResp := &extension.ServiceCreationResponse{}
				cfApi.On("Provision", "source_app_id", request).Return(createAppResp, nil)

				sut := New(dataCatalog, cfApi, nats, CreationStatusFactory{})
				resp, _ := sut.CreateService(request)

				cfApi.AssertExpectations(GinkgoT())
				Expect(resp).NotTo(BeNil())
			})
		})

		Context("with nats configured", func() {
			It("should publish events", func() {
				nats = new(messagebus.MessageBusMock)
				nats.(*messagebus.MessageBusMock).On("Publish", mock.Anything).Return()

				dataCatalog.On("Find", mock.Anything).Return(&extension.ServiceExtension{})
				dataCatalog.On("AppendInstance", mock.Anything).Return()

				request := &cf.ServiceCreationRequest{}
				request.Parameters = make(map[string]string)
				cfApi := new(CfMock)
				createAppResp := &extension.ServiceCreationResponse{}
				cfApi.On("Provision", "", request).Return(createAppResp, nil)

				sut := New(dataCatalog, cfApi, nats, CreationStatusFactory{})
				sut.CreateService(request)

				nats.(*messagebus.MessageBusMock).AssertNumberOfCalls(GinkgoT(), "Publish", 2)
			})
		})
	})

	Describe("delete service", func() {
		Context("in case of cloud foundry error", func() {
			It("should propagate error", func() {
				svcExt := &extension.ServiceInstanceExtension{
					App: types.CfAppResource{Meta: types.CfMeta{GUID: "appGuid"}}}
				dataCatalog.On("FindInstance", mock.Anything).Return(svcExt)

				cfApi := new(CfMock)
				expectedErr := errors.New("ERROR!")
				cfApi.On("Deprovision", mock.Anything).Return(expectedErr)

				sut := New(dataCatalog, cfApi, nats, CreationStatusFactory{})
				err := sut.DeleteService("serviceID")

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when deprovisioning succeeds", func() {
			It("should return non empty response", func() {
				svcExt := &extension.ServiceInstanceExtension{
					ID:  "entryId",
					App: types.CfAppResource{Meta: types.CfMeta{GUID: "appGuid"}}}
				dataCatalog.On("FindInstance", mock.Anything).Return(svcExt)
				dataCatalog.On("RemoveInstance", mock.Anything).Return(nil)

				cfApi := new(CfMock)
				cfApi.On("Deprovision", mock.Anything).Return(nil)

				sut := New(dataCatalog, cfApi, nats, CreationStatusFactory{})
				err := sut.DeleteService("serviceID")

				Expect(err).To(BeNil())
				dataCatalog.AssertCalled(GinkgoT(), "RemoveInstance", "entryId")
			})
		})
	})
})
