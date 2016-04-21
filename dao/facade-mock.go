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

package dao

import (
	"github.com/stretchr/testify/mock"
	"github.com/trustedanalytics/application-broker/service/extension"
)

type FacadeMock struct {
	mock.Mock
}

func (c *FacadeMock) Get() ([]*extension.ServiceExtension, error) {
	args := c.Called()
	return args.Get(0).([]*extension.ServiceExtension), nil
}

func (c *FacadeMock) Append(service *extension.ServiceExtension) (err error) {
	c.Called(service)
	return nil
}

func (c *FacadeMock) Update(service *extension.ServiceExtension) (err error) {
	c.Called(service)
	return nil
}

func (c *FacadeMock) Remove(serviceID string) (err error) {
	c.Called(serviceID)
	return nil
}

func (c *FacadeMock) Find(id string) (*extension.ServiceExtension, error) {
	args := c.Called(id)
	if args.Get(0) == nil { //first return value is nil, we test error case then
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*extension.ServiceExtension), nil
}

func (c *FacadeMock) AppendInstance(instance extension.ServiceInstanceExtension) error {
	c.Called(instance)
	return nil
}

func (c *FacadeMock) FindInstance(id string) (*extension.ServiceInstanceExtension, error) {
	args := c.Called(id)
	return args.Get(0).(*extension.ServiceInstanceExtension), nil
}

func (c *FacadeMock) HasInstancesOf(serviceID string) (bool, error) {
	args := c.Called(serviceID)
	if args.Get(1) != nil { //first return value is nil, we test error case then
		return false, args.Get(1).(error)
	}
	return args.Bool(0), nil
}

func (c *FacadeMock) RemoveInstance(id string) error {
	c.Called(id)
	return nil
}
