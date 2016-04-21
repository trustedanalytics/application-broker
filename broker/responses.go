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
	cf "github.com/cloudfoundry-community/types-cf"
	"github.com/trustedanalytics/application-broker/service/extension"
	"github.com/trustedanalytics/go-cf-lib/types"
)

// OK
// swagger:response emptyBodyOk
type emptyBodyOk struct{}

var emptyOk = emptyBodyOk{}

// Created
// swagger:response emptyBodyCreated
type emptyBodyCreated struct{}

var emptyCreated = emptyBodyCreated{}

// No Content
// swagger:response emptyBodyNoContent
type emptyBodyNoContent struct{}

var emptyNoContent = emptyBodyNoContent{}

// Bad Request
// swagger:response emptyBodyBadRequest
type emptyBodyBadRequest struct{}

var emptyBadRequest = emptyBodyBadRequest{}

// Not Found
// swagger:response emptyBodyNotFound
type emptyBodyNotFound struct{}

var emptyNotFound = emptyBodyNotFound{}

// Conflict
// swagger:response emptyBodyConflict
type emptyBodyConflict struct{}

var emptyConflict = emptyBodyConflict{}

// Internal Server Error
// swagger:response brokerErrorResponse
type BrokerErrorResponse struct {
	// Error description
	// in: body
	Body cf.BrokerError
}

// ServiceCreationResponse
// swagger:response serviceCreationResponse
type ServiceCreationResponse struct {
	// in: body
	Body cf.ServiceCreationResponse
}

// ServiceExtensionResponse
// swagger:response serviceExtensionResponse
type ServiceExtensionResponse struct {
	// in: body
	Body extension.ServiceExtension
}

// CatalogExtensionResponse
// swagger:response catalogExtensionResponse
type CatalogExtensionResponse struct {
	// in: body
	Body extension.CatalogExtension
}

// ServiceBindingResponse
// swagger:response serviceBindingResponse
type ServiceBindingResponse struct {
	// in: body
	Body types.ServiceBindingResponse
}

// swagger:parameters updateService deleteService
type ServiceIdParam struct {
	// Service GUID
	// in: path
	// required: true
	ServiceId string `json:"service_id"`
}

// swagger:parameters provisionServiceInstance deprovisionServiceInstance bindService unbindService
type InstanceIdParam struct {
	// Service instance GUID
	// in: path
	// required: true
	InstanceId string `json:"instance_id"`
}

// swagger:parameters bindService unbindService
type BindingIdParam struct {
	// Service binding GUID
	// in: path
	// required: true
	BindingId string `json:"binding_id"`
}
