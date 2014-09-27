package main

import (
	"fmt"
	"github.com/intel-data/cf-catalog"
	"log"
)

// ServiceProvider object
type SimpleServiceProvider struct {
	DashboardRootURL string
}

// Initialize configures the service provider
func (p *SimpleServiceProvider) Initialize() error {
	log.Println("initializing...")
	// TODO: Load the source of service data here
	p.DashboardRootURL = "https://somename.gotapaas.com"
	return nil
}

// GetServiceDashboard gets service pointer for this id
func (p *SimpleServiceProvider) GetServiceDashboard(id string) (*catalog.CFServiceProvisioningResponse, error) {
	log.Printf("getting service: %s", id)

	d := &catalog.CFServiceProvisioningResponse{}

	// TODO: everything will have to be derived from the source of services
	d.DashboardURL = fmt.Sprintf("%s/dashboard", p.DashboardRootURL)

	return d, nil
}
