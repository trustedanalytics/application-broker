package broker

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cloudfoundry-community/types-cf"
	"github.com/gorilla/mux"
)

// will hold the empty repose "{}"
var empty = struct{}{}

type handler struct {
	provider cf.ServiceProvider
}

func newHandler(p cf.ServiceProvider) *handler {
	return &handler{p}
}

func (h *handler) catalog(r *http.Request) responseEntity {
	log.Println("handler requesting catalog")
	cat, err := h.provider.GetCatalog()
	if err != nil {
		return handleServiceError(err)
	}
	log.Println("handler retrieved catalog")
	return responseEntity{http.StatusOK, cat}
}

func (h *handler) provision(req *http.Request) responseEntity {
	vars := mux.Vars(req)
	preq := &cf.ServiceCreationRequest{InstanceID: vars[instanceID]}
	log.Printf("handler provisioning: %v", preq)
	if err := json.NewDecoder(req.Body).Decode(&preq); err != nil {
		handleDecodingError(err)
	}
	log.Printf("handler provisioning request decoded: %v", preq)
	resp, err := h.provider.CreateService(preq)
	if err != nil {
		return handleServiceError(err)
	}
	log.Printf("handler request provisioned: %v", resp)
	return responseEntity{http.StatusCreated, resp}
}

func (h *handler) deprovision(req *http.Request) responseEntity {
	vars := mux.Vars(req)
	instID := vars[instanceID]
	log.Printf("handler de-provisioning: %s", instID)
	if err := h.provider.DeleteService(instID); err != nil {
		return handleServiceError(err)
	}
	log.Printf("handler de-provisioned: %v", instID)
	return responseEntity{http.StatusOK, empty}
}

func (h *handler) bind(req *http.Request) responseEntity {
	vars := mux.Vars(req)
	breq := &cf.ServiceBindingRequest{
		InstanceID: vars[instanceID],
		BindingID:  vars[bindingID],
	}
	log.Printf("handler binding: %v", breq)
	if err := json.NewDecoder(req.Body).Decode(&breq); err != nil {
		handleDecodingError(err)
	}
	log.Printf("handler binding request decoded: %v", breq)
	resp, err := h.provider.BindService(breq)
	if err != nil {
		return handleServiceError(err)
	}
	log.Printf("handler bound: %v", resp)
	return responseEntity{http.StatusCreated, resp}
}

func (h *handler) unbind(req *http.Request) responseEntity {
	vars := mux.Vars(req)
	instID := vars[instanceID]
	bindID := vars[bindingID]
	log.Printf("handler unbinding: %s for %s", bindID, instID)
	if err := h.provider.UnbindService(instID, bindID); err != nil {
		return handleServiceError(err)
	}
	log.Printf("handler unbound: %s for %s", bindID, instID)
	return responseEntity{http.StatusOK, empty}
}

// helpers
func handleDecodingError(err error) responseEntity {
	log.Printf("decoding error: %v", err)
	return responseEntity{
		http.StatusBadRequest,
		cf.BrokerError{Description: err.Error()},
	}
}

func handleServiceError(err *cf.ServiceProviderError) responseEntity {
	log.Printf("handler service error: %v", err)
	if err == nil {
		return responseEntity{http.StatusInternalServerError, empty}
	}
	log.Fatalf("internal server error: %s", err.String())
	switch err.Code {
	case cf.ErrorInstanceExists:
		return responseEntity{http.StatusConflict, empty}
	case cf.ErrorInstanceNotFound:
		return responseEntity{http.StatusGone, empty}
	default:
		return responseEntity{
			http.StatusInternalServerError,
			cf.BrokerError{Description: err.String()},
		}
	}
}
