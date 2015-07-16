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
	"log"

	"github.com/intel-data/app-launching-service-broker/messagebus"
	"github.com/cloudfoundry-community/types-cf"
)

// LaunchingService object
type LaunchingService struct {
	config *ServiceConfig
	client *CFClient
	nats   messagebus.MessageBus
}

// New creates an isntance of the LaunchingService
func New(natsInstance messagebus.MessageBus) (*LaunchingService, error) {
	s := &LaunchingService{
		config: Config,
		client: NewCFClient(Config),
		nats:   natsInstance,
	}
	return s, nil
}

// GetCatalog parses catalog response
func (p *LaunchingService) GetCatalog() (*cf.Catalog, *cf.ServiceProviderError) {
	log.Println("getting catalog...")
	return p.config.Catalog, nil
}

// CreateService create a service instance
func (p *LaunchingService) CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponse, *cf.ServiceProviderError) {
	log.Printf("creating service: %v", r)
	msg := &ServiceCreationStatus{ServiceId: r.InstanceID, ServiceType: Config.ServiceName, Message: "CreateService operation started"}
	p.nats.Publish(msg)

	dashboardUrl := ""
	if p.config.DashboardURL != "" {
		dashboardUrl = p.config.DashboardURL + "/" + r.InstanceID
	}
	d := &cf.ServiceCreationResponse{DashboardURL: dashboardUrl}

	ctx, err := p.client.getContextFromSpaceOrg(r.InstanceID, r.SpaceGUID, r.OrganizationGUID)
	if err != nil {
		log.Printf("error getting app: %v", err)
		msg.Message = "Getting context failed while service creation. Err: " + err.Error()
		p.nats.Publish(msg)
		return nil, cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	err = p.client.provision(ctx)
	if err != nil {
		msg.Message = "Service spawning failed with error: " + err.Error()
	} else {
		msg.Message = "Service spawning succeded"
	}
	p.nats.Publish(msg)
	return d, nil
}

// DeleteService deletes itself and its dependencies
func (p *LaunchingService) DeleteService(instanceID string) *cf.ServiceProviderError {
	log.Printf("deleting service: %s", instanceID)

	ctx, err := p.client.getContextFromServiceInstanceID(instanceID)
	if err != nil {
		log.Printf("error getting app: %v", err)
		return cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	err = p.client.deprovision(ctx)
	if err != nil {
		cf.NewServiceProviderError(cf.ErrorServerException, err)
	}

	return nil
}

// BindService creates a service instance binding
func (p *LaunchingService) BindService(r *cf.ServiceBindingRequest) (*cf.ServiceBindingResponse, *cf.ServiceProviderError) {
	log.Printf("creating service binding: %v", r)

	b := &cf.ServiceBindingResponse{}

	ctx, err := p.client.getContextFromServiceInstanceID(r.InstanceID)
	if err != nil {
		log.Printf("error getting service: %v", err)
		return nil, cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	app, err := p.client.getAppByName(ctx.SpaceGUID, ctx.AppName)
	if err != nil {
		log.Printf("error getting app by name: %v", err)
		return nil, cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	// TODO: See if the above is even needed for this generic kind of an app
	log.Printf("binding - ctx[%v] app[%v]", ctx, app)

	// TODO: Return app URL from the context in API
	b.Credentials = make(map[string]string)

	// TODO: Set this to the app URI
	// func to get first route for AppName
	route, err := p.client.getFirstFullRouteURL(app)
	if err != nil {
		log.Printf("error getting app route: %v", err)
		return nil, cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}
	log.Printf("app route - %s", route)

	appCreds, err := p.client.runSetupScript(ctx)
	if err != nil {
		return nil, cf.NewServiceProviderError(cf.ErrorServerException, err)
	}
	for key, value := range appCreds {
		b.Credentials[key] = value
	}

	b.Credentials["name"] = ctx.AppName
	b.Credentials["route"] = route
	b.Credentials["url"] = "https://" + route // TODO determine protocol

	return b, nil
}

// UnbindService deletes service instance binding
func (p *LaunchingService) UnbindService(instanceID, bindingID string) *cf.ServiceProviderError {
	log.Printf("deleting service binding: %s/%s", instanceID, bindingID)

	// NOTE: Currently no action required for unbinding; return Ok

	// ctx, err := p.client.getContextFromServiceInstanceID(instanceID)
	// if err != nil {
	// 	log.Printf("error getting service: %v", err)
	// 	return cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	// }
	//
	// bind, err := p.client.getBinding(bindingID)
	// if err != nil {
	// 	log.Printf("error getting binding: %v", err)
	// 	return cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	// }
	//
	// app, err := p.client.getApp(bind.AppGUID)
	// if err != nil {
	// 	log.Printf("error getting app: %v", err)
	// 	return cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	// }
	//
	// // TODO: See if the above is even needed for this generic kind of an app
	// log.Printf("binding - bind[%v] ctx[%v] app[%v]", bind, ctx, app)

	return nil
}
