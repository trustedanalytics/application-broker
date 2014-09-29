package broker

import (
	"github.com/emicklei/go-restful"
	"github.com/intel-data/types-cf"
	"log"
	"net/http"
)

// ServiceProvider defines the required provider functionality
type ServiceProvider interface {
	createService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponce, error)
	deleteService(id string) error
}

// ServiceBindingProvider defines the required provider functionality
type ServiceBindingProvider interface {
	bindService(r *cf.ServiceBindingRequest, serviceID, bindingID string) (*cf.ServiceBindingResponse, error)
	unbindService(serviceID, bindingID string) error
}

// ServiceHandler object
type ServiceHandler struct {
	serviceProvider        ServiceProvider
	serviceBindingProvider ServiceBindingProvider
}

func (h *ServiceHandler) initialize() error {
	log.Println("initializing...")

	// TODO: Load the provider, is there a IOC pattern in go?

	s := &SimpleServiceProvider{}
	s.initialize()
	h.serviceProvider = s

	b := &SimpleServiceBindingProvider{}
	b.initialize()
	h.serviceBindingProvider = b

	return nil
}

func (h *ServiceHandler) createService(request *restful.Request, response *restful.Response) {
	if !hasRequiredParams(request, response, "serviceId") {
		return
	}

	id := request.PathParameter("serviceId")
	log.Printf("getting service instance for id: %s", id)

	// marshal request
	req := &cf.ServiceCreationRequest{}
	err := request.ReadEntity(req)
	if err != nil {
		handleServerError(response, err)
		return
	}

	// get service dashboard
	d, err := h.serviceProvider.createService(req)
	if err != nil {
		handleServerError(response, err)
		return
	}

	/*
	   201 created
	   409 already exists
	   200 nothing changed
	*/

	log.Printf("service created: %d", http.StatusCreated)
	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(d)

}

func (h *ServiceHandler) createServiceBinding(request *restful.Request, response *restful.Response) {
	if !hasRequiredParams(request, response, "serviceID", "bindingID") {
		return
	}

	serviceID := request.PathParameter("serviceID")
	bindingID := request.PathParameter("bindingID")
	log.Printf("creating binding %s/%s", serviceID, bindingID)

	// parse request
	req := &cf.ServiceBindingRequest{}
	err := request.ReadEntity(req)
	if err != nil {
		handleServerError(response, err)
		return
	}

	// build response
	res, err := h.serviceBindingProvider.bindService(req, serviceID, bindingID)
	if err != nil {
		handleServerError(response, err)
		return
	}

	/*
	   201 created
	   409 already exists
	   200 nothing changed
	*/

	log.Printf("service binding created: %d", http.StatusCreated)
	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(res)

}

func (h *ServiceHandler) deleteServiceBinding(request *restful.Request, response *restful.Response) {
	if !hasRequiredParams(request, response, "serviceID", "bindingID") {
		return
	}

	serviceID := request.PathParameter("serviceID")
	bindingID := request.PathParameter("bindingID")
	log.Printf("deleting binding %s/%s", serviceID, bindingID)

	err := h.serviceBindingProvider.unbindService(serviceID, bindingID)
	if err != nil {
		handleServerError(response, err)
		return
	}

	log.Printf("service instance binding has been deleted: %d", http.StatusGone)
	response.WriteHeader(http.StatusGone)

	/*
	   201 created
	   410 gone
	*/

}

func (h *ServiceHandler) deleteService(request *restful.Request, response *restful.Response) {
	if !hasRequiredParams(request, response, "serviceID") {
		return
	}

	serviceID := request.PathParameter("serviceID")
	log.Printf("deleting service: %s", serviceID)

	err := h.serviceProvider.deleteService(serviceID)
	if err != nil {
		handleServerError(response, err)
		return
	}

	log.Printf("service instance has been deleted: %d", http.StatusGone)
	response.WriteHeader(http.StatusGone)

	/*
	   201 created
	   410 gone
	*/

}
