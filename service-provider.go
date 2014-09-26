package main

import (
	catalog "github.com/intel-data/cf-catalog"
	"log"
)

// ServiceProvider object
type ServiceProvider struct {
}

// Initialize configures the service provider
func (p *ServiceProvider) Initialize() error {
	log.Println("initializing...")
	// TODO: Load the source of service data here
	return nil
}

// GetServiceInstance gets service pointer for this id
func (p *ServiceProvider) GetServiceInstance(id string) (*catalog.CFServiceState, error) {
	log.Printf("getting service: %s", id)

	s := &catalog.CFServiceState{}

	// TODO: everything will have to be derived from the source of services

	s.ServiceID = catalog.NewID()
	s.PlanID = catalog.NewID()
	s.OrganizationGUID = catalog.NewID()
	s.SpaceGUID = catalog.NewID()

	return s, nil
}
