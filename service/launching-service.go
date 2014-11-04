package service

import (
	"log"

	"github.com/intel-data/types-cf"
)

// LaunchingService object
type LaunchingService struct {
	config *ServiceConfig
	client *CFClient
}

// New creates an isntance of the LaunchingService
func New() (*LaunchingService, error) {
	s := &LaunchingService{
		config: Config,
		client: NewCFClient(Config),
	}
	return s, nil
}

// GetCatalog parses catalog response
func (p *LaunchingService) GetCatalog() (*cf.Catalog, *cf.ServiceProviderError) {
	log.Println("getting catalog...")
	return p.config.Catalog, nil
}

// CreateService create a service instance
func (p *LaunchingService) CreateService(r *cf.ServiceCreationRequest) (*cf.ServiceCreationResponce, *cf.ServiceProviderError) {
	log.Printf("creating service: %v", r)
	d := &cf.ServiceCreationResponce{}

	ctx, err := p.client.getContext(r.InstanceID)
	if err != nil {
		log.Printf("error getting app: %v", err)
		return nil, cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	p.client.provision(ctx)

	// TODO: Validate this is the correct URI to return
	d.DashboardURL = ctx.ServiceURI

	return d, nil
}

// DeleteService deletes itself and its dependencies
func (p *LaunchingService) DeleteService(instanceID string) *cf.ServiceProviderError {
	log.Printf("deleting service: %s", instanceID)

	ctx, err := p.client.getContext(instanceID)
	if err != nil {
		log.Printf("error getting app: %v", err)
		return cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	p.client.deprovision(ctx)

	return nil
}

// BindService creates a service instance binding
func (p *LaunchingService) BindService(r *cf.ServiceBindingRequest) (*cf.ServiceBindingResponse, *cf.ServiceProviderError) {
	log.Printf("creating service binding: %v", r)

	b := &cf.ServiceBindingResponse{}

	ctx, err := p.client.getContext(r.InstanceID)
	if err != nil {
		log.Printf("error getting service: %v", err)
		return nil, cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	app, err := p.client.getApp(r.AppGUID)
	if err != nil {
		log.Printf("error getting app: %v", err)
		return nil, cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	// TODO: See if the above is even needed for this generic kind of an app
	log.Printf("binding - ctx[%v] app[%v]", ctx, app)

	// TODO: Return app URL from the context in API
	b.Credentials = make(map[string]string)

	// TODO: Set this to the app URI
	b.Credentials["URI"] = "mysql://user:pass@localhost:9999/" + ctx.ServiceName

	return b, nil
}

// UnbindService deletes service instance binding
func (p *LaunchingService) UnbindService(instanceID, bindingID string) *cf.ServiceProviderError {
	log.Printf("deleting service binding: %s/%s", instanceID, bindingID)

	ctx, err := p.client.getContext(instanceID)
	if err != nil {
		log.Printf("error getting service: %v", err)
		return cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	bind, err := p.client.getBinding(bindingID)
	if err != nil {
		log.Printf("error getting binding: %v", err)
		return cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	app, err := p.client.getApp(bind.AppGUID)
	if err != nil {
		log.Printf("error getting app: %v", err)
		return cf.NewServiceProviderError(cf.ErrorInstanceNotFound, err)
	}

	// TODO: See if the above is even needed for this generic kind of an app
	log.Printf("binding - bind[%v] ctx[%v] app[%v]", bind, ctx, app)

	return nil
}
