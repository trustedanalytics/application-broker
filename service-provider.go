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

// SetServiceBinding sets service binding for a service
func (p *SimpleServiceProvider) SetServiceBinding(instanceId, serviceId string) (*catalog.CFServiceBindingResponse, error) {
	log.Printf("setting service binding: %s/%s", instanceId, serviceId)

	b := &catalog.CFServiceBindingResponse{}

	b.Credentials = &catalog.CFServiceCredential{}
	b.Credentials.URI = "mysql://user:pass@localhost:3306/dbname"
	b.Credentials.Hostname = "localhost"
	b.Credentials.Port = "3306"
	b.Credentials.Name = "dbname"
	b.Credentials.Vhost = "amqp://yser:pass@host/queue"
	b.Credentials.Username = "user"
	b.Credentials.Password = "pass"
	b.SyslogDrainURL = "syslog://logs.example.com"

	return b, nil
}

// DeleteServiceBinding delete service binding for a service
func (p *SimpleServiceProvider) DeleteServiceBinding(instanceId, serviceId string) error {
	log.Printf("setting service binding: %s/%s", instanceId, serviceId)

	// TODO: implement

	return nil
}
