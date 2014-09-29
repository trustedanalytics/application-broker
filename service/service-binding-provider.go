package service

import (
	"github.com/intel-data/generic-cf-service-broker/common"
	"github.com/intel-data/types-cf"
	"log"
)

// SimpleServiceBindingProvider object
type SimpleServiceBindingProvider struct{}

// Initialize configures the service provider
func (p *SimpleServiceBindingProvider) Initialize() error {
	log.Println("initializing...")
	return nil
}

// BindService creates a service instance binding
func (p *SimpleServiceBindingProvider) BindService(r *cf.ServiceBindingRequest, serviceID, bindingID string) (*cf.ServiceBindingResponse, *common.ServiceProviderError) {
	log.Printf("creating service binding: %v - %s/%s", r, serviceID, bindingID)

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
func (p *SimpleServiceBindingProvider) UnbindService(serviceID, bindingID string) *common.ServiceProviderError {
	log.Printf("deleting service binding: %s/%s", serviceID, bindingID)
	// TODO: implement
	return nil
}
