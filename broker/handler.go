package broker

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cloudfoundry-community/types-cf"
	"github.com/go-martini/martini"
)

// will hold the empty repose "{}"
var empty = struct{}{}

type handler struct {
	provider cf.ServiceProvider
}

func newHandler(p cf.ServiceProvider) *handler {
	return &handler{p}
}

func (h *handler) catalog(r *http.Request, params martini.Params) (int, string) {
	log.Println("handler requesting catalog")
	cat, err := h.provider.GetCatalog()
	if err != nil {
		return handleServiceError(err)
	}
	log.Println("handler retrieved catalog")
	return marshalEntity(responseEntity{http.StatusOK, cat})
}

func (h *handler) provision(req *http.Request, params martini.Params) (int, string) {
	preq := &cf.ServiceCreationRequest{InstanceID: params["instance_id"]}
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
	return marshalEntity(responseEntity{http.StatusCreated, resp})
}

func (h *handler) deprovision(req *http.Request, params martini.Params) (int, string) {
	instID := params["instance_id"]
	log.Printf("handler de-provisioning: %s", instID)
	if err := h.provider.DeleteService(instID); err != nil {
		return handleServiceError(err)
	}
	log.Printf("handler de-provisioned: %v", instID)
	return marshalEntity(responseEntity{http.StatusOK, empty})
}

func (h *handler) bind(req *http.Request, params martini.Params) (int, string) {
	breq := &cf.ServiceBindingRequest{
		InstanceID: params["instance_id"],
		BindingID:  params["binding_id"],
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
	return marshalEntity(responseEntity{http.StatusCreated, resp})
}

func (h *handler) unbind(req *http.Request, params martini.Params) (int, string) {
	instID := params["instance_id"]
	bindID := params["binding_id"]
	log.Printf("handler unbinding: %s for %s", bindID, instID)
	if err := h.provider.UnbindService(instID, bindID); err != nil {
		return handleServiceError(err)
	}
	log.Printf("handler unbound: %s for %s", bindID, instID)
	return marshalEntity(responseEntity{http.StatusOK, empty})
}

// helpers
func handleDecodingError(err error) (int, string) {
	log.Printf("decoding error: %v", err)
	return marshalEntity(responseEntity{
		http.StatusBadRequest,
		cf.BrokerError{Description: err.Error()},
	})
}

func handleServiceError(err *cf.ServiceProviderError) (int, string) {
	log.Printf("handler service error: %v", err)
	if err == nil {
		return marshalEntity(responseEntity{http.StatusInternalServerError, empty})
	}
	log.Fatalf("internal server error: %s", err.String())
	switch err.Code {
	case cf.ErrorInstanceExists:
		return marshalEntity(responseEntity{http.StatusConflict, empty})
	case cf.ErrorInstanceNotFound:
		return marshalEntity(responseEntity{http.StatusGone, empty})
	case cf.ErrorServerException:
		return marshalEntity(responseEntity{http.StatusInternalServerError, empty})
	default:
		return marshalEntity(responseEntity{
			http.StatusInternalServerError,
			cf.BrokerError{Description: err.String()},
		})
	}
}

func marshalEntity(entity responseEntity) (int, string) {
	payload, err := json.Marshal(entity.value)
	if err != nil {
		log.Fatalf("internal server error: %s", err)
		return 500, ""
	}
	return entity.status, string(payload)
}
