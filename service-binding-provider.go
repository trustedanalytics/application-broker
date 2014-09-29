package main

import (
	"github.com/intel-data/types-cf"
	"log"
)

// SimpleServiceBindingProvider object
type SimpleServiceBindingProvider struct{}

// Initialize configures the service provider
func (p *SimpleServiceBindingProvider) initialize() error {
	log.Println("initializing...")
	return nil
}

func (p *SimpleServiceBindingProvider) bindService(r *cf.ServiceBindingRequest, serviceID, bindingID string) (*cf.ServiceBindingResponse, error) {
	log.Printf("creating service binding: %v - %s/%s", r, serviceID, bindingID)
	b := &cf.ServiceBindingResponse{}
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

func (p *SimpleServiceBindingProvider) unbindService(serviceID, bindingID string) error {
	log.Printf("deleting service binding: %s/%s", serviceID, bindingID)
	// TODO: implement
	return nil
}
