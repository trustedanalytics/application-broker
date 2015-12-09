/**
 * Copyright (c) 2015 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package broker

import (
	"encoding/json"
	"net/http"

	log "github.com/cihub/seelog"

	"github.com/cloudfoundry-community/types-cf"
	"github.com/go-martini/martini"
	"github.com/nu7hatch/gouuid"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
)

// will hold the empty reponse "{}"
var empty = struct{}{}

type handler struct {
	provider types.ServiceProviderExtension
}

func newHandler(p types.ServiceProviderExtension) *handler {
	return &handler{p}
}

func (h *handler) append(req *http.Request, params martini.Params) (int, string) {
	log.Infof("handler appending new service to catalog: [%v]", req.Body)
	toAdd := types.NewAutogeneratedService()
	err := json.NewDecoder(req.Body).Decode(&toAdd)
	if err != nil {
		return handleDecodingError(err)
	}
	log.Debugf("handler provisioning request decoded: %+v", toAdd)

	if err := h.provider.InsertToCatalog(toAdd); err != nil {
		return handleServiceError(err)
	}
	return marshalEntity(responseEntity{http.StatusCreated, toAdd})
}

func (h *handler) update(req *http.Request, params martini.Params) (int, string) {
	service_id := params["service_id"]
	log.Infof("handler updating service id: [%v] in catalog: [%v]", service_id, req.Body)
	toUpdate := new(types.ServiceExtension)

	err := json.NewDecoder(req.Body).Decode(&toUpdate)
	if err != nil {
		return handleDecodingError(err)
	}
	if service_id != toUpdate.ID {
		log.Warn("Service id in URL different from service id in body send")
		return handleServiceError(misc.InvalidInputError{})
	}
	log.Debugf("handler provisioning update decoded: %+v", toUpdate)

	if err := h.provider.UpdateCatalog(toUpdate); err != nil {
		return handleServiceError(err)
	}
	log.Infof("ID: %v", toUpdate.ID)
	return marshalEntity(responseEntity{http.StatusOK, toUpdate})
}

func (h *handler) remove(req *http.Request, params martini.Params) (int, string) {
	log.Info("handler removing service from catalog")
	err := h.provider.DeleteFromCatalog(params["service_id"])
	if err != nil {
		return handleServiceError(err)
	}
	return marshalEntity(responseEntity{http.StatusNoContent, empty})
}

func (h *handler) catalog(r *http.Request, params martini.Params) (int, string) {
	log.Info("handler requesting catalog")
	catalog, err := h.provider.GetCatalog()
	if err != nil {
		return handleServiceError(err)
	}
	log.Debug("handler retrieved catalog")
	return marshalEntity(responseEntity{http.StatusOK, catalog})
}

func (h *handler) provision(req *http.Request, params martini.Params) (int, string) {
	preq := &cf.ServiceCreationRequest{InstanceID: params["instance_id"]}
	if err := json.NewDecoder(req.Body).Decode(&preq); err != nil {
		return handleDecodingError(err)
	}
	if preq.Parameters == nil {
		preq.Parameters = map[string]string{}
	}
	if preq.Parameters["name"] == "" {
		random, _ := uuid.NewV4()
		preq.Parameters["name"] = random.String()
	}
	log.Infof("handler provisioning: [%v]", preq.Parameters["name"])
	log.Debugf("handler provisioning request decoded: [%+v]", preq)
	resp, err := h.provider.CreateService(preq)
	if err != nil {
		return handleServiceError(err)
	}
	log.Debugf("handler request provisioned - response: [%+v]", resp)
	return marshalEntity(responseEntity{http.StatusCreated, resp})
}

func (h *handler) deprovision(req *http.Request, params martini.Params) (int, string) {
	instID := params["instance_id"]
	log.Infof("handler de-provisioning: %s", instID)
	if err := h.provider.DeleteService(instID); err != nil {
		return handleServiceError(err)
	}
	log.Debugf("handler de-provisioned: %v", instID)
	return marshalEntity(responseEntity{http.StatusOK, empty})
}

func (h *handler) bind(req *http.Request, params martini.Params) (int, string) {
	breq := &cf.ServiceBindingRequest{
		InstanceID: params["instance_id"],
		BindingID:  params["binding_id"],
	}
	log.Infof("handler binding: %v", breq)
	if err := json.NewDecoder(req.Body).Decode(&breq); err != nil {
		handleDecodingError(err)
	}
	log.Debugf("handler binding request decoded: %v", breq)
	resp, err := h.provider.BindService(breq)
	if err != nil {
		return handleServiceError(err)
	}
	log.Debugf("handler bound: %v", resp)
	status := http.StatusCreated
	if breq.AppGUID == "" {
		status = http.StatusOK
	}
	return marshalEntity(responseEntity{status, resp})
}

func (h *handler) unbind(req *http.Request, params martini.Params) (int, string) {
	instID := params["instance_id"]
	bindID := params["binding_id"]
	log.Infof("handler unbinding: %s for %s", bindID, instID)
	// NOTE: Currently no action required for unbinding; return Ok
	return marshalEntity(responseEntity{http.StatusOK, empty})
}

// helpers
func handleDecodingError(err error) (int, string) {
	log.Errorf("decoding error: %v", err)
	return marshalEntity(responseEntity{
		http.StatusBadRequest,
		cf.BrokerError{Description: err.Error()},
	})
}

func handleServiceError(err error) (int, string) {
	log.Errorf("handler service error: %v", err)
	switch err {
	case misc.ServiceAlreadyExistsError{}:
		return marshalEntity(responseEntity{http.StatusConflict, empty})
	case misc.InvalidInputError{}:
		return marshalEntity(responseEntity{http.StatusBadRequest, empty})
	case misc.InstanceNotFoundError{}, misc.ServiceNotFoundError{}:
		return marshalEntity(responseEntity{http.StatusNotFound, empty})
	case misc.InternalServerError{}:
		return marshalEntity(responseEntity{http.StatusInternalServerError, empty})
	default:
		return marshalEntity(responseEntity{
			http.StatusInternalServerError,
			cf.BrokerError{Description: err.Error()},
		})
	}
}

func marshalEntity(entity responseEntity) (int, string) {
	payload, err := json.Marshal(entity.value)
	if err != nil {
		log.Errorf("internal server error: %s", err)
		return 500, ""
	}
	return entity.status, string(payload)
}
