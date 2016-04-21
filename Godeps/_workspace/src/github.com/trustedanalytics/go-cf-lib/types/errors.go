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

import "github.com/juju/errors"

var InvalidInputError error = errors.New("Invalid request body")
var InstanceAlreadyExistsError = errors.New("Such an instance already exists")
var ServiceAlreadyExistsError = errors.New("Service already exists")
var ServiceNotFoundError = errors.New("No such service exists")
var InstanceNotFoundError = errors.New("No such instance exists")
var EntityNotFoundError = errors.New("Entity does not exist")
var InternalServerError = errors.New("Some internal error occurred")
var CcJobFailedError = errors.New("Error occurred while copying bits")
var CcRestageFailedError = errors.New("Error occurred while restaging")
var CcUpdateFailedError = errors.New("Error occurred while app updating")
var CcGetInstancesFailedError = errors.New("Error occurred while getting app instances")
var TimeoutOccurredError = errors.New("Asynchronous call timeouted")
var ExistingInstancesError = errors.New("Can't remove service with existing instances from catalog")
