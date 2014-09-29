package main

import (
	"fmt"
	"github.com/intel-data/types-cf"
	"log"
)

// ServiceProvider object
type SimpleServiceProvider struct {
	DashboardRootURL string
}

func (p *SimpleServiceProvider) initialize() error {
	log.Println("initializing...")
	// TODO: Load the source of service data here
	p.DashboardRootURL = "https://somename.gotapaas.com"
	return nil
}

func (p *SimpleServiceProvider) createService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponce, error) {
	log.Printf("creating service: %v", r)
	d := &cf.ServiceCreationResponce{}
	// TODO: everything will have to be derived from the source of services
	d.DashboardURL = fmt.Sprintf("%s/dashboard", p.DashboardRootURL)
	return d, nil
}

func (p *SimpleServiceProvider) deleteService(id string) error {
	log.Printf("deleting service: %s", id)
	// TODO: implement
	return nil
}
