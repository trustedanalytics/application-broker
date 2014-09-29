package broker

import (
	"github.com/intel-data/types-cf"
	"log"
)

const (
	AppID          = "3427569C-2A11-456C-974C-106B221E5EB2"
	AppVersion     = "0.1.0"
	AppName        = "generic-cf-service-broker"
	AppDescription = "Dynamically configurable service broker"
)

// CatalogProvider object
type MockedCatalogProvider struct{}

// Initialize configures the catalog provider
func (p *MockedCatalogProvider) initialize() error {
	log.Println("initializing...")
	// TODO: Load the source of catalog data here
	return nil
}

// TODO: fix the return types to standard object, error when implemented
func (p *MockedCatalogProvider) newSerivcePlan(id, name, desc string) *cf.Plan {
	log.Printf("creating service plan: %s", id)
	pl := &cf.Plan{}
	pl.ID = id
	pl.Name = name
	pl.Description = desc
	pl.Free = true
	return pl
}

func (p *MockedCatalogProvider) newSerivce(id string) (*cf.Service, error) {
	log.Printf("creating service: %s", id)
	s := &cf.Service{}
	// TODO: everything will have to be derived from the source of services
	s.ID = id
	s.Name = AppName
	s.Description = AppDescription
	s.Bindable = true
	s.Tags = []string{"generic", "service", "broker"}
	s.Plans = []*cf.Plan{
		p.newSerivcePlan(s.ID+"-1", s.Name+"-1", s.Description+"-1"),
		p.newSerivcePlan(s.ID+"-2", s.Name+"-2", s.Description+"-2"),
		p.newSerivcePlan(s.ID+"-3", s.Name+"-3", s.Description+"-3"),
	}
	return s, nil
}

func (p *MockedCatalogProvider) getCatalog() (*cf.Catalog, error) {
	log.Println("getting catalog...")
	c := &cf.Catalog{}
	// TODO: query service store and generate these on the fly

	// downside of embedding in go is that you no longer can just
	// {ID:"123", Name:"abc", Desc: "some"}
	s, err := p.newSerivce(AppID)
	if err != nil {
		log.Printf("Error while making service: %v", err)
		return nil, err
	}
	s.Dashboard = &cf.Dashboard{
		ID:     s.ID + "-9",
		Secret: "secret",
		URI:    "http://dashboard.host.com/d",
	}
	c.Services = []*cf.Service{s}
	return c, nil
}
