package main

import (
	"github.com/intel-data/cf-catalog"
	"log"
)

// CatalogProvider object
type MockedCatalogProvider struct {
}

// Initialize configures the catalog provider
func (p *MockedCatalogProvider) Initialize() error {
	log.Println("initializing...")
	// TODO: Load the source of catalog data here
	return nil
}

// TODO: fix the return types to standard object, error when implemented
func (p *MockedCatalogProvider) newSerivcePlan(id, name, desc string) *catalog.CFPlan {
	log.Printf("making service plan: %s", id)
	pl := &catalog.CFPlan{}
	pl.ID = id
	pl.Name = name
	pl.Description = desc
	pl.Free = true
	return pl
}

func (p *MockedCatalogProvider) newSerivce(id string) (*catalog.CFService, error) {
	log.Printf("making service: %s", id)
	s := &catalog.CFService{}
	// TODO: everything will have to be derived from the source of services
	s.ID = id
	s.Name = "generic-cf-service-broker"
	s.Description = "Dynamically configurable service broker"
	s.Bindable = false
	s.Tags = []string{"generic", "service", "broker"}
	s.Plans = []*catalog.CFPlan{
		p.newSerivcePlan(s.ID+"-1", s.Name+"-1", s.Description+"-1"),
		p.newSerivcePlan(s.ID+"-2", s.Name+"-2", s.Description+"-2"),
		p.newSerivcePlan(s.ID+"-3", s.Name+"-3", s.Description+"-3"),
	}
	return s, nil
}

// GetCatalog gets the catalog
func (p *MockedCatalogProvider) GetCatalog() (*catalog.CFCatalog, error) {
	log.Println("getting catalog...")
	c := &catalog.CFCatalog{}
	// TODO: query service store and generate these on the fly

	// downside of embedding in go is that you no longer can just
	// {ID:"123", Name:"abc", Desc: "some"}
	s1ID := "3427569C-2A11-456C-974C-106B221E5EB2"
	s1, err := p.newSerivce(s1ID)
	if err != nil {
		log.Printf("Error while making service: %s", s1ID)
		return nil, err
	}

	s2ID := "458F5495-FB29-4127-A9E6-370E2F20670A"
	s2, err := p.newSerivce(s1ID)
	if err != nil {
		log.Printf("Error while making service: %s", s2ID)
		return nil, err
	}

	c.Services = []*catalog.CFService{s1, s2}

	return c, nil
}
