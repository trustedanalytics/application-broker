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
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/cloudfoundry-community/types-cf"
	"github.com/trustedanalytics/application-broker/cloud"
	"github.com/trustedanalytics/application-broker/dao"
	"github.com/trustedanalytics/application-broker/messagebus"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
)

// LaunchingService wraps access to db, cloud controller and messagebus
type LaunchingService struct {
	db         dao.Facade
	cloud      cloud.API
	msgBus     messagebus.MessageBus
	msgFactory messagebus.MessageFactory
}

// New creates an instance of the LaunchingService
func New(db dao.Facade, cloud cloud.API, natsInstance messagebus.MessageBus, messageFactory messagebus.MessageFactory) *LaunchingService {
	s := &LaunchingService{
		db:         db,
		cloud:      cloud,
		msgBus:     natsInstance,
		msgFactory: messageFactory,
	}
	return s
}

// InsertToCatalog adds new application description that can be spawned/duplicated on demand
// Description is stored in underlying implementation of Catalog interface
func (p *LaunchingService) InsertToCatalog(svc *types.ServiceExtension) error {
	if !types.Validate(svc) {
		return misc.InvalidInputError{}
	}
	if err := p.db.Append(svc); err != nil {
		return err
	}

	if err := p.UpdateBroker(); err != nil {
		return err
	}

	return nil
}

// UpdateCatalog update application description that can be spawned/duplicated on demand
// Description is stored in underlying implementation of Catalog interface
func (p *LaunchingService) UpdateCatalog(svc *types.ServiceExtension) error {
	if !types.Validate(svc) {
		return misc.InvalidInputError{}
	}
	if err := p.db.Update(svc); err != nil {
		return err
	}

	if err := p.UpdateBroker(); err != nil {
		return err
	}

	return nil
}

// DeleteFromCatalog deletes data pointing to reference application from internal storage
func (p *LaunchingService) DeleteFromCatalog(serviceID string) error {
	if _, err := p.db.Find(serviceID); err != nil {
		return err
	}

	services, err := p.db.Get()
	if err != nil {
		return err
	}
	if len(services) <= 1 {
		return misc.InternalServerError{Context: "Cannot delete the only service offering. Catalog cannot be empty."}
	}

	hasInstances, err := p.db.HasInstancesOf(serviceID)
	if err != nil {
		return err
	}

	if hasInstances {
		return misc.ExistingInstancesError{}
	}

	if err := p.db.Remove(serviceID); err != nil {
		return err
	}

	if err := p.UpdateBroker(); err != nil {
		return err
	}

	return nil
}

// GetCatalog parses catalog response
func (p *LaunchingService) GetCatalog() (*types.CatalogExtension, error) {
	log.Debug("getting catalog...")

	var err error
	toReturn := types.CatalogExtension{}
	toReturn.Services, err = p.db.Get()

	if err != nil {
		return nil, err
	}
	return &toReturn, nil
}

// CreateService creates a service instance
func (p *LaunchingService) CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponse, error) {
	service, err := p.db.Find(r.ServiceID)
	if err != nil {
		return nil, err
	}

	if r.Parameters["name"] == "" {
		r.Parameters["name"] = service.Name
	}

	idx := strings.Index(r.InstanceID, "-")
	if idx > 0 {
		// Take only first part of the GUID
		r.Parameters["name"] = r.Parameters["name"] + "-" + r.InstanceID[0:idx]
	}

	name := r.Parameters["name"]
	log.Infof("create service: [%v]", name)

	stype := service.Name
	org := r.OrganizationGUID

	msg := p.msgFactory.NewServiceStatus(name, stype, org, "CreateService operation started")
	p.msgBus.Publish(msg)

	//TODO: instead of referenceApp.GUID we should pass entire app object
	resp, err := p.cloud.Provision(service.ReferenceApp.Meta.GUID, r)
	if err != nil {
		msg = p.msgFactory.NewServiceStatus(name, stype, org, "Service spawning failed with error: "+err.Error())
		p.msgBus.Publish(msg)
		return nil, err
	}
	msg = p.msgFactory.NewServiceStatus(name, stype, org, "Service spawning succeded")
	p.msgBus.Publish(msg)

	toAppend := types.ServiceInstanceExtension{
		App:       resp.App,
		ID:        r.InstanceID,
		ServiceID: r.ServiceID,
	}
	if err := p.db.AppendInstance(toAppend); err != nil {
		return nil, err
	}
	return &resp.ServiceCreationResponse, nil
}

// DeleteService deletes service instance and its dependencies
func (p *LaunchingService) DeleteService(instanceID string) error {
	log.Debugf("Deleting service %s...", instanceID)

	service, err := p.db.FindInstance(instanceID)
	if err != nil {
		return err
	}

	if err := p.cloud.Deprovision(service.App.Meta.GUID); err != nil {
		return err
	}
	p.db.RemoveInstance(service.ID)
	return nil
}

// BindService creates a (service instance <-> application) binding
func (p *LaunchingService) BindService(r *cf.ServiceBindingRequest) (*types.ServiceBindingResponse, error) {
	instance, err := p.db.FindInstance(r.InstanceID)
	if err != nil {
		return nil, err
	}

	resp := new(types.ServiceBindingResponse)
	resp.Credentials = make(map[string]string)
	resp.Credentials["url"] = instance.App.Meta.URL
	return resp, nil
}

func (p *LaunchingService) UpdateBroker() error {
	vcapSerialized := misc.GetEnvVarAsString("VCAP_APPLICATION", "{}")
	vcap := new(types.CfVcapApplication)
	json.NewDecoder(bytes.NewReader([]byte(vcapSerialized))).Decode(vcap)
	username := misc.GetEnvVarAsString("AUTH_USER", "")
	password := misc.GetEnvVarAsString("AUTH_PASS", "")

	if len(vcap.Name) == 0 {
		return misc.InternalServerError{Context: "Application name is not set"}
	}

	if len(vcap.Uris) == 0 || len(vcap.Uris[0]) == 0 {
		return misc.InternalServerError{Context: "Application has no url set"}
	}
	url := fmt.Sprintf("http://%v", vcap.Uris[0])

	return p.cloud.UpdateBroker(vcap.Name, url, username, password)
}
