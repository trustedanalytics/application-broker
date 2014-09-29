package broker

import (
	"github.com/emicklei/go-restful"
	"github.com/intel-data/generic-cf-service-broker/common"
	"github.com/intel-data/types-cf"
	"log"
	"net/http"
)

// Handler object
type Handler struct {
	provider common.ServiceProvider
}

func NewHandler(p common.ServiceProvider) (*Handler, error) {
	log.Println("initializing...")
	h := &Handler{provider: p}
	return h, nil
}

func (h *Handler) getCatalog(request *restful.Request, response *restful.Response) {
	log.Println("getting catalog...")
	c, err := h.provider.GetCatalog()
	if err != nil {
		handleServerError(response, err)
	} else {
		response.WriteHeader(http.StatusOK)
		response.WriteEntity(c)
	}
}

func (h *Handler) createService(request *restful.Request, response *restful.Response) {
	if !hasRequiredParams(request, response, "serviceId") {
		return
	}

	id := request.PathParameter("serviceId")
	log.Printf("getting service instance for id: %s", id)

	// marshal request
	req := &cf.ServiceCreationRequest{}
	err := request.ReadEntity(req)
	if err != nil {
		handleSimpleServerError(response, err)
		return
	}

	// get service dashboard
	d, err2 := h.provider.CreateService(req)
	if err2 != nil {
		handleServerError(response, err2)
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

func (h *Handler) createServiceBinding(request *restful.Request, response *restful.Response) {
	if !hasRequiredParams(request, response, "serviceId", "bindingId") {
		return
	}

	serviceID := request.PathParameter("serviceId")
	bindingID := request.PathParameter("bindingId")
	log.Printf("creating binding %s/%s", serviceID, bindingID)

	// parse request
	req := &cf.ServiceBindingRequest{}
	err := request.ReadEntity(req)
	if err != nil {
		handleSimpleServerError(response, err)
		return
	}

	// build response
	res, err2 := h.provider.BindService(req, serviceID, bindingID)
	if err2 != nil {
		handleServerError(response, err2)
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

func (h *Handler) deleteServiceBinding(request *restful.Request, response *restful.Response) {
	if !hasRequiredParams(request, response, "serviceId", "bindingId") {
		return
	}

	serviceID := request.PathParameter("serviceId")
	bindingID := request.PathParameter("bindingId")
	log.Printf("deleting binding %s/%s", serviceID, bindingID)

	err := h.provider.UnbindService(serviceID, bindingID)
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

func (h *Handler) deleteService(request *restful.Request, response *restful.Response) {
	if !hasRequiredParams(request, response, "serviceId") {
		return
	}

	serviceID := request.PathParameter("serviceId")
	log.Printf("deleting service: %s", serviceID)

	err := h.provider.DeleteService(serviceID)
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
