package main

import (
	"github.com/emicklei/go-restful"
	"github.com/intel-data/cf-catalog"
	"log"
	"net/http"
)

// ServiceProvider defines the required provider functionality
type ServiceProvider interface {
	Initialize() error
	GetServiceDashboard(id string) (*catalog.CFServiceProvisioningResponse, error)
	SetServiceBinding(instanceId, serviceId string) (*catalog.CFServiceBindingResponse, error)
	DeleteServiceBinding(instanceId, serviceId string) error
}

// ServiceHandler object
type ServiceHandler struct {
	Provider ServiceProvider
}

// Initialize configures the broker handler
func (h *ServiceHandler) Initialize() error {
	log.Println("initializing...")
	// TODO: Load the provider, is there a IOC pattern in go?
	c := &SimpleServiceProvider{}
	c.Initialize()
	h.Provider = c
	return nil
}

// SetServiceInstance returns a list of instances for particular service
func (h *ServiceHandler) SetServiceInstance(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	log.Printf("getting service instance for id: %s", id)

	s := &catalog.CFServiceProvisioningResponse{}
	err := request.ReadEntity(s)

	if err != nil {
		log.Printf("error on parsing service state %s: %v", id, err)
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteErrorString(
			http.StatusInternalServerError,
			"Error resource creation")
		return
	}

	d, err := h.Provider.GetServiceDashboard(id)
	if err != nil {
		log.Printf("error on getting dashboard: %v", err)
		response.WriteErrorString(
			http.StatusInternalServerError,
			"Error creating catalog")
		return
	}

	log.Printf("service instance has been created: %d", http.StatusCreated)
	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(d)

	/*
	   201 created
	   409 already exists
	   200 nothing changed
	*/

}

// SetServiceInstanceBinding returns a list of instances for particular service
func (h *ServiceHandler) SetServiceInstanceBinding(request *restful.Request, response *restful.Response) {
	instanceId := request.PathParameter("instanceId")
	serviceId := request.PathParameter("bindingId")
	log.Printf("setting service instance [%s] binding [%s]", instanceId, serviceId)

	req := &catalog.CFServiceBindingRequest{}
	err := request.ReadEntity(req)

	if err != nil {
		log.Printf("error on parsing service binding request: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteErrorString(
			http.StatusInternalServerError,
			"Error resource creation")
		return
	}

	bind, err := h.Provider.SetServiceBinding(instanceId, serviceId)
	if err != nil {
		log.Printf("error on getting dashboard: %v", err)
		response.WriteErrorString(
			http.StatusInternalServerError,
			"Error creating catalog")
		return
	}

	log.Printf("service instance has been created: %d", http.StatusCreated)
	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(bind)

	/*
	   201 created
	   409 already exists
	   200 nothing changed
	*/

}

// DeleteServiceInstanceBinding deletes instances for particular service
func (h *ServiceHandler) DeleteServiceInstanceBinding(request *restful.Request, response *restful.Response) {
	instanceId := request.PathParameter("instanceId")
	serviceId := request.PathParameter("bindingId")
	log.Printf("setting service instance [%s] binding [%s]", instanceId, serviceId)

	err := h.Provider.DeleteServiceBinding(instanceId, serviceId)
	if err != nil {
		log.Printf("error on getting dashboard: %v", err)
		response.WriteErrorString(
			http.StatusInternalServerError,
			"Error creating catalog")
		return
	}

	log.Printf("service instance has been deleted: %d", http.StatusGone)
	response.WriteHeader(http.StatusGone)

	/*
	   201 created
	   410 gone
	*/

}
