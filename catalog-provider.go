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

// NewSerivce gets service pointer for this id
func (p *MockedCatalogProvider) NewSerivce(id string) (*catalog.CFService, error) {
	log.Printf("making service: %s", id)

	s := &catalog.CFService{}

	// TODO: everything will have to be derived from the source of services
	//       for now just hard-code these

	s.ID = id
	s.Name = "generic-cf-service-broker"
	s.Description = "Dynamically configurable service broker"
	s.Bindable = false
	s.Tags = []string{"generic", "service", "broker"}
	s.Plans = []*catalog.CFPlan{
		{
			ID:          id + "-1",
			Name:        "Service 1",
			Description: "Service 1 description",
			Free:        true,
		}, {
			ID:          id + "-2",
			Name:        "Service 2",
			Description: "Service 2 description",
			Free:        true,
		},
	}

	return s, nil
}

// GetCatalog gets the catalog
func (p *MockedCatalogProvider) GetCatalog() (*catalog.CFCatalog, error) {
	log.Println("getting catalog...")
	c := &catalog.CFCatalog{}

	// TODO: query service store and generate these on the fly
	s1ID := "3427569C-2A11-456C-974C-106B221E5EB2"
	s1, err := p.NewSerivce(s1ID)
	if err != nil {
		log.Printf("Error while making service: %s", s1ID)
		return nil, err
	}

	s2ID := "458F5495-FB29-4127-A9E6-370E2F20670A"
	s2, err := p.NewSerivce(s1ID)
	if err != nil {
		log.Printf("Error while making service: %s", s2ID)
		return nil, err
	}

	c.Services = []*catalog.CFService{s1, s2}

	return c, nil
}
