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

import "github.com/cloudfoundry-community/types-cf"

// ServiceProviderExtension beside standard cf.ServiceProvider introduces additional API endpoints.
type ServiceProviderExtension interface {

	// Appends service to the catalog managed by this broker
	InsertToCatalog(*ServiceExtension) error

	// Updates service in the catalog managed by this broker
	UpdateCatalog(*ServiceExtension) error

	// Deletes service description from the catalog
	DeleteFromCatalog(serviceID string) error

	// GetCatalog returns the catalog of services managed by this broker
	GetCatalog() (*CatalogExtension, error)

	// CreateService creates a service instance for specific plan
	CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponse, error)

	// DeleteService deletes previously created service instance
	DeleteService(instanceID string) error

	// BindService binds to specified service instance and
	// Returns credentials necessary to establish connection to that service
	BindService(r *cf.ServiceBindingRequest) (*ServiceBindingResponse, error)
}
