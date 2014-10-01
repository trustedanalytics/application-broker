package broker

import (
	sr "github.com/emicklei/go-restful"
	"github.com/intel-data/types-cf"
	"log"
	"net/http"
)

// Handler object
type Handler struct {
	provider cf.ServiceProvider
}

func NewHandler(p cf.ServiceProvider) (*Handler, error) {
	log.Println("initializing...")
	h := &Handler{provider: p}
	return h, nil
}

func (h *Handler) getCatalog(request *sr.Request, response *sr.Response) {
	log.Println("getting catalog...")
	c, err := h.provider.GetCatalog()
	if err != nil {
		handleServerError(response, err)
	} else {
		response.WriteHeader(http.StatusOK)
		response.WriteEntity(c)
	}
}

func (h *Handler) createService(request *sr.Request, response *sr.Response) {
	if !hasRequiredParams(request, response, "instanceId") {
		return
	}

	instanceId := request.PathParameter("instanceId")
	log.Printf("getting service instance for id: %s", instanceId)

	// marshal request
	req := &cf.ServiceCreationRequest{}
	err := request.ReadEntity(req)
	if err != nil {
		handleSimpleServerError(response, err)
		return
	}

	// add path args to req obj
	req.InstanceID = instanceId

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

func (h *Handler) createServiceBinding(request *sr.Request, response *sr.Response) {
	if !hasRequiredParams(request, response, "instanceId", "bindingId") {
		return
	}

	instanceId := request.PathParameter("instanceId")
	bindingId := request.PathParameter("bindingId")
	log.Printf("creating binding %s/%s", instanceId, bindingId)

	// parse request
	req := &cf.ServiceBindingRequest{}
	err := request.ReadEntity(req)
	if err != nil {
		handleSimpleServerError(response, err)
		return
	}

	// add path args to req obj
	req.InstanceID = instanceId
	req.BindingID = bindingId

	// build response
	res, err2 := h.provider.BindService(req)
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

func (h *Handler) deleteServiceBinding(request *sr.Request, response *sr.Response) {
	if !hasRequiredParams(request, response, "instanceId", "bindingId") {
		return
	}

	instanceId := request.PathParameter("instanceId")
	bindingId := request.PathParameter("bindingId")
	log.Printf("deleting binding %s/%s", instanceId, bindingId)

	err := h.provider.UnbindService(instanceId, bindingId)
	if err != nil {
		handleServerError(response, err)
		return
	}

	log.Printf("service instance binding has been deleted: %d", http.StatusGone)
	response.WriteHeader(http.StatusGone)
	response.WriteAsJson("{}")

	/*
	   201 created
	   410 gone
	*/

}

func (h *Handler) deleteService(request *sr.Request, response *sr.Response) {
	if !hasRequiredParams(request, response, "instanceId") {
		return
	}

	instanceId := request.PathParameter("instanceId")
	log.Printf("deleting service instance: %s", instanceId)

	err := h.provider.DeleteService(instanceId)
	if err != nil {
		handleServerError(response, err)
		return
	}

	log.Printf("service instance has been deleted: %d", http.StatusGone)
	response.WriteHeader(http.StatusGone)
	response.WriteAsJson("{}")

	/*
	   201 created
	   410 gone
	*/

}
