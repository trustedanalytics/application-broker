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
	"github.com/cloudfoundry-community/types-cf"
	"github.com/stretchr/testify/mock"
	"github.com/trustedanalytics/application-broker/service/extension"
)

type CfMock struct {
	mock.Mock
}

func (c *CfMock) Provision(sourceAppGUID string, request *cf.ServiceCreationRequest) (*extension.ServiceCreationResponse, error) {
	args := c.Called(sourceAppGUID, request)
	if args.Get(0) == nil {
		//first return value is nil, we test error case then
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*extension.ServiceCreationResponse), nil
}

func (c *CfMock) Deprovision(appGUID string) error {
	args := c.Called(appGUID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (c *CfMock) UpdateBroker(brokerName string, brokerUri string, username string, password string) error {
	args := c.Called(brokerName, brokerUri, username, password)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (c *CfMock) CheckIfServiceExists(serviceName string) error {
	c.Called(serviceName)
	return nil
}
