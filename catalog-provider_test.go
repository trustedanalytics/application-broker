package main

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestGetCatalog(t *testing.T) {

	p := &MockedCatalogProvider{}
	p.Initialize()

	catalog, err := p.GetCatalog()

	assert.Nil(t, err, err)
	assert.NotNil(t, catalog, "nil catalog")
	assert.NotNil(t, catalog.Services, "nil catalog services")

	for i, srv := range catalog.Services {
		log.Printf("service:%d - %v ", i, srv)

		// check the required fields
		assert.NotEmpty(t, srv.ID, "nil service ID")
		assert.NotEmpty(t, srv.Name, "nil service name")
		assert.NotEmpty(t, srv.Description, "nil service description")
		assert.NotNil(t, srv.Plans, "nil service plans")

		for j, pln := range srv.Plans {
			log.Printf("service plan:%d - %v ", j, pln)

			// check the required fields
			assert.NotEmpty(t, pln.ID, "nil plan ID")
			assert.NotEmpty(t, pln.Name, "nil plan name")
			assert.NotEmpty(t, pln.Description, "nil plan description")
			assert.NotNil(t, pln.Free, "nil plan free indicator")

		}

	}

}
