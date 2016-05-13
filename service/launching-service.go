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
	"fmt"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/cloudfoundry-community/types-cf"
	"github.com/juju/errors"
	"github.com/trustedanalytics/application-broker/cloud"
	"github.com/trustedanalytics/application-broker/dao"
	"github.com/trustedanalytics/application-broker/env"
	"github.com/trustedanalytics/application-broker/messagebus"
	"github.com/trustedanalytics/application-broker/service/extension"
	"github.com/trustedanalytics/go-cf-lib/types"
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
func (p *LaunchingService) InsertToCatalog(svc *extension.ServiceExtension) error {
	if !extension.Validate(svc) {
		return types.InvalidInputError
	}

	if err := p.db.Append(svc); err != nil {
		return err
	}

	if err := p.cloud.CheckIfServiceExists(svc.Name); err != nil {
		p.db.Remove(svc.ID)
		return err
	}
	if err := p.UpdateBroker(); err != nil {
		p.db.Remove(svc.ID)
		return err
	}

	return nil
}

// UpdateCatalog update application description that can be spawned/duplicated on demand
// Description is stored in underlying implementation of Catalog interface
func (p *LaunchingService) UpdateCatalog(svc *extension.ServiceExtension) error {
	if !extension.Validate(svc) {
		return types.InvalidInputError
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
		return errors.Annotate(types.InternalServerError, "Cannot delete the only service offering. Catalog cannot be empty.")
	}

	hasInstances, err := p.db.HasInstancesOf(serviceID)
	if err != nil {
		return err
	}

	if hasInstances {
		return types.ExistingInstancesError
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
func (p *LaunchingService) GetCatalog() (*extension.CatalogExtension, error) {
	log.Debug("getting catalog...")

	var err error
	toReturn := extension.CatalogExtension{}
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

	r.Parameters["name"] = p.normalizeInstanceName(r.InstanceID, r.Parameters["name"], service.Name)
	name := r.Parameters["name"]
	log.Infof("create service: [%v]", name)

	stype := service.Name
	org := r.OrganizationGUID

	msg := p.msgFactory.NewServiceStatus(name, stype, org, "CreateService operation started")
	p.msgBus.Publish(msg)

	//TODO: instead of referenceApp.GUID we should pass entire app object
	resp, err := p.cloud.Provision(service.ReferenceApp.Meta.GUID, service.Configuration, r)
	if err != nil {
		msg = p.msgFactory.NewServiceStatus(name, stype, org, "Service spawning failed with error: "+err.Error())
		p.msgBus.Publish(msg)

		if appendErr := p.appendInstance(r, resp); appendErr != nil {
			log.Errorf("Failed to append instance %v to database: [%v]", r.InstanceID, appendErr.Error())
		}
		return nil, err
	}
	msg = p.msgFactory.NewServiceStatus(name, stype, org, "Service spawning succeded")
	p.msgBus.Publish(msg)

	if appendErr := p.appendInstance(r, resp); appendErr != nil {
		log.Errorf("Failed to append instance %v to database: [%v]", r.InstanceID, appendErr.Error())
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
	vcap := env.GetVcapApplication()
	username := env.GetEnvVarAsString("AUTH_USER", "")
	password := env.GetEnvVarAsString("AUTH_PASS", "")

	if len(vcap.Name) == 0 {
		return errors.Annotate(types.InternalServerError, "Application name is not set")
	}

	if len(vcap.Uris) == 0 || len(vcap.Uris[0]) == 0 {
		return errors.Annotate(types.InternalServerError, "Application has no url set")
	}
	url := fmt.Sprintf("http://%v", vcap.Uris[0])

	return p.cloud.UpdateBroker(vcap.Name, url, username, password)
}

func (p *LaunchingService) appendInstance(req *cf.ServiceCreationRequest, res *extension.ServiceCreationResponse) error {
	toAppend := extension.ServiceInstanceExtension{
		ID:        req.InstanceID,
		ServiceID: req.ServiceID,
	}
	if res != nil {
		toAppend.App = res.App
	}
	return p.db.AppendInstance(toAppend)
}

func (p *LaunchingService) normalizeInstanceName(instanceID string, instanceName string, serviceName string) (string) {
	nameToNormalize := getNameToNormalize(instanceName, serviceName)
	nameToNormalize = replaceSpacesByDashes(nameToNormalize)
	normalizedInstanceName := addInstanceIdSuffix(instanceID, nameToNormalize)
	return normalizedInstanceName
}

func getNameToNormalize(instanceName string, serviceName string) (string) {
	if (instanceName == "") {
		return serviceName
	}
	return instanceName
}

func replaceSpacesByDashes(name string) (string) {
	return strings.Replace(name, " ", "-", -1)
}

func addInstanceIdSuffix(instanceID string, instanceName string) (string) {
	// Suffix required for ATK
	idx := strings.Index(instanceID, "-")
	if idx > 0 {
		// Take only first part of the GUID
		instanceName = instanceName + "-" + instanceID[0:idx]
	}
	return instanceName
}
