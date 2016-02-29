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

package types

type InvalidInputError struct{}

func (i InvalidInputError) Error() string {
	return "Invalid request body"
}

type InstanceAlreadyExistsError struct{}

func (i InstanceAlreadyExistsError) Error() string {
	return "Such an instance already exists"
}

type ServiceAlreadyExistsError struct{}

func (i ServiceAlreadyExistsError) Error() string {
	return "Service already exists"
}

type ServiceNotFoundError struct{}

func (i ServiceNotFoundError) Error() string {
	return "No such service exists"
}

type InstanceNotFoundError struct{}

func (i InstanceNotFoundError) Error() string {
	return "No such instance exists"
}

type EntityNotFoundError struct{}

func (i EntityNotFoundError) Error() string {
	return "Entity does not exist"
}

type InternalServerError struct {
	Context string
}

func (i InternalServerError) Error() string {
	return "Some internal error occurred: " + i.Context
}

type CcJobFailedError struct {
	InternalCfMessage string
}

func (i CcJobFailedError) Error() string {
	return "Error occurred while copying bits: " + i.InternalCfMessage
}

type CcRestageFailedError struct {
	InternalCfMessage string
}

func (i CcRestageFailedError) Error() string {
	return "Error occurred while restaging: " + i.InternalCfMessage
}

type CcUpdateFailedError struct {
	InternalCfMessage string
}

func (i CcUpdateFailedError) Error() string {
	return "Error occurred while app updating: " + i.InternalCfMessage
}

type CcGetInstancesFailedError struct {
	InternalCfMessage string
}

func (i CcGetInstancesFailedError) Error() string {
	return "Error occurred while getting app instances: " + i.InternalCfMessage
}

type TimeoutOccurredError struct{}

func (i TimeoutOccurredError) Error() string {
	return "Asynchronous call timeouted"
}

type ExistingInstancesError struct{}

func (i ExistingInstancesError) Error() string {
	return "Can't remove service with existing instances from catalog"
}
