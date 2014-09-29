package common

import (
	"github.com/intel-data/types-cf"
)

// ServiceProvider defines the required provider functionality
type ServiceProvider interface {
	GetVersion() string
	GetCatalog() (*cf.Catalog, *ServiceProviderError)
	CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponce, *ServiceProviderError)
	DeleteService(id string) *ServiceProviderError
	BindService(r *cf.ServiceBindingRequest, serviceID, bindingID string) (*cf.ServiceBindingResponse, *ServiceProviderError)
	UnbindService(serviceID, bindingID string) *ServiceProviderError
}
