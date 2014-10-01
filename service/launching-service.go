package service

import (
	"fmt"
	"github.com/intel-data/types-cf"
	"log"
)

// LaunchingService object
type LaunchingService struct {
	config *ServiceConfig
	client *CFClient
}

func New() (*LaunchingService, error) {
	s := &LaunchingService{
		config: Config,
		client: NewCFClient(Config),
	}
	return s, nil
}

// GetCatalog parses catalog response
func (p *LaunchingService) GetCatalog() (*cf.Catalog, *cf.ServiceProviderError) {
	log.Println("getting catalog...")
	return p.config.Catalog, nil
}

// CreateService create a service instance
func (p *LaunchingService) CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponce, *cf.ServiceProviderError) {
	log.Printf("creating service: %v", r)
	d := &cf.ServiceCreationResponce{}

	// TODO: Get plan/org - r.SpaceGUID

	// TODO: Get service name - r.ServiceID

	//p.client.provision(app, org, space)

	// TODO: implement
	d.DashboardURL = fmt.Sprintf("http://%s:%d/dashboard", p.config.CFEnv.Host, p.config.CFEnv.Port)
	return d, nil
}

// DeleteService deletes a service instance
func (p *LaunchingService) DeleteService(instanceID string) *cf.ServiceProviderError {
	log.Printf("deleting service: %s", instanceID)
	// TODO: implement
	return nil
}

// BindService creates a service instance binding
func (p *LaunchingService) BindService(r *cf.ServiceBindingRequest) (*cf.ServiceBindingResponse, *cf.ServiceProviderError) {
	log.Printf("creating service binding: %v", r)

	b := &cf.ServiceBindingResponse{}

	// TODO: implement the service binding logic here
	b.Credentials = &cf.Credential{}
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

// UnbindService deletes service instance binding
func (p *LaunchingService) UnbindService(instanceID, bindingID string) *cf.ServiceProviderError {
	log.Printf("deleting service binding: %s/%s", instanceID, bindingID)
	// TODO: implement
	return nil
}
