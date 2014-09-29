package service

import (
	"fmt"
	"github.com/intel-data/generic-cf-service-broker/common"
	"github.com/intel-data/types-cf"
	"log"
)

// ServiceProvider object
type SimpleServiceProvider struct {
	DashboardRootURL string
}

func (p *SimpleServiceProvider) Initialize() error {
	log.Println("initializing...")
	// TODO: Load the source of service data here
	p.DashboardRootURL = "https://somename.gotapaas.com"
	return nil
}

// CreateService create a service instance
func (p *SimpleServiceProvider) CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponce, *common.ServiceProviderError) {
	log.Printf("creating service: %v", r)
	d := &cf.ServiceCreationResponce{}
	// TODO: implement
	d.DashboardURL = fmt.Sprintf("%s/dashboard", p.DashboardRootURL)
	return d, nil
}

// DeleteService deletes a service instance
func (p *SimpleServiceProvider) DeleteService(id string) *common.ServiceProviderError {
	log.Printf("deleting service: %s", id)
	// TODO: implement
	return nil
}
