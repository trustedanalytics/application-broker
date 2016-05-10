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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/trustedanalytics/application-broker/service/extension"
	"strings"
)

var _ = Describe("Internals", func() {

	Describe("select accepted service params", func() {
		var (
			sut                      *CloudAPI
			passedParams             map[string]string
			allServicesConfiguration []*extension.ServiceConfiguration
		)

		BeforeEach(func() {
			sut = NewCloudAPI(nil)
			passedParams = make(map[string]string)
			allServicesConfiguration = []*extension.ServiceConfiguration{}
		})

		serviceName := "hdfs"

		Context("passed empty params and configuration", func() {
			It("should return null", func() {

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, passedParams, allServicesConfiguration)

				Expect(acceptedParams).Should(BeNil())
			})
		})

		Context("passed nil params", func() {
			It("should return null", func() {

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, nil, allServicesConfiguration)

				Expect(acceptedParams).Should(BeNil())
			})
		})

		Context("passed nil configuration", func() {
			It("should return null", func() {

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, passedParams, nil)

				Expect(acceptedParams).Should(BeNil())
			})
		})

		Context("passed nil params and configuration", func() {
			It("should return null", func() {

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, nil, nil)

				Expect(acceptedParams).Should(BeNil())
			})
		})

		Context("passed configuration for different service only", func() {
			It("should return null", func() {
				passedParams["key"] = "value"
				serviceConf := extension.ServiceConfiguration{
					ServiceName: "other",
					Params:      []string{"otherKey"},
				}
				allServicesConfiguration = append(allServicesConfiguration, &serviceConf)

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, passedParams, allServicesConfiguration)

				Expect(acceptedParams).Should(BeNil())
			})
		})

		Context("passed configurable key without namespace", func() {
			It("should pass key", func() {
				passedParams["key"] = "value"
				serviceConf := extension.ServiceConfiguration{
					ServiceName: serviceName,
					Params:      []string{"key"},
				}
				allServicesConfiguration = append(allServicesConfiguration, &serviceConf)

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, passedParams, allServicesConfiguration)

				Expect(acceptedParams["key"]).Should(Equal("value"))
			})
		})

		Context("passed configurable key with namespace", func() {
			It("should pass key", func() {
				passedParams[serviceName+".key"] = "value"
				serviceConf := extension.ServiceConfiguration{
					ServiceName: serviceName,
					Params:      []string{"key"},
				}
				allServicesConfiguration = append(allServicesConfiguration, &serviceConf)

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, passedParams, allServicesConfiguration)

				Expect(acceptedParams["key"]).Should(Equal("value"))
			})
		})

		Context("passed key which is configurable for different service", func() {
			It("should return nil", func() {
				passedParams[serviceName+".key"] = "value"
				serviceConf := extension.ServiceConfiguration{
					ServiceName: serviceName,
					Params:      []string{"different"},
				}
				otherServiceConf := extension.ServiceConfiguration{
					ServiceName: "other",
					Params:      []string{"key"},
				}
				allServicesConfiguration = append(allServicesConfiguration, &otherServiceConf)
				allServicesConfiguration = append(allServicesConfiguration, &serviceConf)

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, passedParams, allServicesConfiguration)

				Expect(acceptedParams["key"]).Should(BeNil())
			})
		})

		Context("passed key which is configurable for other service", func() {
			It("should return nil", func() {
				passedParams["different.key"] = "value"
				serviceConf := extension.ServiceConfiguration{
					ServiceName: serviceName,
					Params:      []string{"different"},
				}
				otherServiceConf := extension.ServiceConfiguration{
					ServiceName: "other",
					Params:      []string{"key"},
				}
				allServicesConfiguration = append(allServicesConfiguration, &otherServiceConf)
				allServicesConfiguration = append(allServicesConfiguration, &serviceConf)

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, passedParams, allServicesConfiguration)

				Expect(acceptedParams["key"]).Should(BeNil())
			})
		})

		Context("passed key which is configurable for different service only", func() {
			It("should return nil", func() {
				passedParams[serviceName+".key"] = "value"
				serviceConf := extension.ServiceConfiguration{
					ServiceName: "different",
					Params:      []string{"key"},
				}

				allServicesConfiguration = append(allServicesConfiguration, &serviceConf)

				acceptedParams := sut.selectAcceptedServiceParams(serviceName, passedParams, allServicesConfiguration)

				Expect(acceptedParams["key"]).Should(BeNil())
			})
		})
	})

	Describe("remove parameters namespaces", func() {
		var (
			sut          *CloudAPI
			passedParams map[string]string
		)

		BeforeEach(func() {
			sut = NewCloudAPI(nil)
			passedParams = make(map[string]string)
		})

		Context("passed empty params", func() {
			It("should return empty map", func() {

				acceptedParams, err := sut.removeParametersNamespaces(passedParams)

				Expect(err).Should(BeNil())
				Expect(len(acceptedParams)).Should(Equal(0))
			})
		})

		Context("passed nil params", func() {
			It("should return nil", func() {

				acceptedParams, err := sut.removeParametersNamespaces(nil)

				Expect(err).Should(BeNil())
				Expect(acceptedParams).Should(BeNil())
			})
		})

		Context("passed params without namespace", func() {
			It("should be passed as is", func() {
				passedParams["key"] = "value"
				passedParams["key2"] = "value2"

				acceptedParams, err := sut.removeParametersNamespaces(passedParams)

				Expect(err).Should(BeNil())
				Expect(acceptedParams["key"]).Should(Equal(passedParams["key"]))
				Expect(acceptedParams["key2"]).Should(Equal(passedParams["key2"]))
			})
		})

		Context("passed params with namespace", func() {
			It("only keys should be passed", func() {
				passedParams["namespace.key"] = "value"
				passedParams["service.key2"] = "value2"

				acceptedParams, err := sut.removeParametersNamespaces(passedParams)

				Expect(err).Should(BeNil())
				Expect(acceptedParams["key"]).Should(Equal(passedParams["namespace.key"]))
				Expect(acceptedParams["key2"]).Should(Equal(passedParams["service.key2"]))
			})
		})

		Context("passed same param with and without namespace", func() {
			It("should return error with info about colision for specific key", func() {
				passedParams["key"] = "value"
				passedParams["service.key"] = "value2"

				acceptedParams, err := sut.removeParametersNamespaces(passedParams)

				Expect(err).ShouldNot(BeNil())
				Expect(strings.Contains(err.Error(), "key")).Should(Equal(true))
				Expect(acceptedParams).Should(BeNil())
			})
		})

		Context("passed same param without and with namespace", func() {
			It("should return error with info about colision for specific key", func() {
				passedParams["service.key"] = "value2"
				passedParams["key"] = "value"

				acceptedParams, err := sut.removeParametersNamespaces(passedParams)

				Expect(err).ShouldNot(BeNil())
				Expect(strings.Contains(err.Error(), "key")).Should(Equal(true))
				Expect(acceptedParams).Should(BeNil())
			})
		})

	})

})
