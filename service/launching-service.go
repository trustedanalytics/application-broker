package service

import (
	"fmt"
	"github.com/intel-data/types-cf"
	"log"
)

const (
	AppID          = "3427569C-2A11-456C-974C-106B221E5EB2"
	AppVersion     = "0.1.0"
	AppName        = "generic-cf-service-broker"
	AppDescription = "Dynamically configurable service broker"
)

// LaunchingService object
type LaunchingService struct {
	config *Config
}

func New() (*LaunchingService, error) {
	s := &LaunchingService{
		config: ServiceConfig,
	}
	return s, nil
}

// GetVersion returns the service version
func (p *LaunchingService) GetVersion() string {
	return AppVersion
}

// GetCatalog parses catalog response
func (p *LaunchingService) GetCatalog() (*cf.Catalog, *cf.ServiceProviderError) {
	log.Println("getting catalog...")
	c := &cf.Catalog{}
	// TODO: implement the service creation logic here

	// downside of embedding in go is that you no longer can just
	// {ID:"123", Name:"abc", Desc: "some"}
	s, err := p.newSerivce(AppID)
	if err != nil {
		log.Printf("Error while making service: %v", err)
		return nil, cf.NewServiceProviderError(cf.ErrorServerException, err)
	}
	s.Dashboard = &cf.Dashboard{
		ID:     s.ID + "-9",
		Secret: "secret",
		URI:    "http://dashboard.host.com/d",
	}
	c.Services = []*cf.Service{s}
	return c, nil
}

// CreateService create a service instance
func (p *LaunchingService) CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponce, *cf.ServiceProviderError) {
	log.Printf("creating service: %v", r)
	d := &cf.ServiceCreationResponce{}
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

// TODO: fix the return types to standard object, error when implemented
func (p *LaunchingService) newSerivcePlan(id, name, desc string) *cf.Plan {
	log.Printf("creating service plan: %s", id)
	pl := &cf.Plan{}
	pl.ID = id
	pl.Name = name
	pl.Description = desc
	pl.Free = true
	return pl
}

func (p *LaunchingService) newSerivce(id string) (*cf.Service, error) {
	log.Printf("creating service: %s", id)
	s := &cf.Service{}
	// TODO: everything will have to be derived from the source of services
	s.ID = id
	s.Name = AppName
	s.Description = AppDescription
	s.Bindable = true
	s.Tags = []string{"generic", "service", "broker"}
	s.Plans = []*cf.Plan{
		p.newSerivcePlan(s.ID+"-1", s.Name+"-1", s.Description+"-1"),
		p.newSerivcePlan(s.ID+"-2", s.Name+"-2", s.Description+"-2"),
		p.newSerivcePlan(s.ID+"-3", s.Name+"-3", s.Description+"-3"),
	}
	return s, nil
}
