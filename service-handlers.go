package main

import (
	"github.com/emicklei/go-restful"
	catalog "github.com/intel-data/cf-catalog"
	"log"
	"net/http"
)

// ServiceHandler object
type ServiceHandler struct {
	Provider *ServiceProvider
}

// Initialize configures the broker handler
func (h *ServiceHandler) Initialize() error {
	log.Println("initializing...")
	// TODO: Load the provider, is there a IOC pattern in go?
	c := &ServiceProvider{}
	c.Initialize()
	h.Provider = c
	return nil
}

// GetInstances returns a list of instances for particular service
func (h *ServiceHandler) GetInstances(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	log.Printf("getting service instance for id: %s", id)

	s := &catalog.CFServiceState{}
	err := request.ReadEntity(s)

	if err != nil {
		log.Printf("error on parsing service state %s: %v", id, err)
		response.WriteErrorString(
			http.StatusInternalServerError,
			"Error resource creation")
	} else {
		log.Printf("service instance has been created: %d", http.StatusCreated)
		response.WriteHeader(http.StatusCreated)
	}
}
