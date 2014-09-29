package common

import (
	"github.com/intel-data/types-cf"
)

// CatalogProvider defines the required provider functionality
type CatalogProvider interface {
	GetCatalog() (*cf.Catalog, *ServiceProviderError)
}

// ServiceProvider defines the required provider functionality
type ServiceProvider interface {
	CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponce, *ServiceProviderError)
	DeleteService(id string) *ServiceProviderError
}

// ServiceBindingProvider defines the required provider functionality
type ServiceBindingProvider interface {
	BindService(r *cf.ServiceBindingRequest, serviceID, bindingID string) (*cf.ServiceBindingResponse, *ServiceProviderError)
	UnbindService(serviceID, bindingID string) *ServiceProviderError
}
