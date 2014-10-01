package service

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

const (
	TestServiceId = "291E642F-E786-49EF-B6FC-F8AF96A36A37"
)

func TestGetCatalog(t *testing.T) {

	p, err := New()
	assert.Nil(t, err, "error on create")
	assert.NotNil(t, p, "nil provider")

	catalog, err2 := p.GetCatalog()

	assert.Nil(t, err2, err2)
	assert.NotNil(t, catalog, "nil catalog")
	assert.NotNil(t, catalog.Services, "nil catalog services")

	for i, srv := range catalog.Services {
		log.Printf("service:%d", i)

		// check the required fields
		assert.NotEmpty(t, srv.ID, "nil service ID")
		assert.NotEmpty(t, srv.Name, "nil service name")
		assert.NotEmpty(t, srv.Description, "nil service description")
		assert.NotNil(t, srv.Plans, "nil service plans")

		if srv.Dashboard != nil {
			log.Printf("dashboard: %d", i)
			assert.NotNil(t, srv.Dashboard.ID, "nil services dashboard id")
			assert.NotNil(t, srv.Dashboard.Secret, "nil services dashboard secret")
			assert.NotNil(t, srv.Dashboard.URI, "nil services dashboard URL")
		}

		for j, pln := range srv.Plans {
			log.Printf("service plan:%d[%d]", i, j)

			// check the required fields
			assert.NotEmpty(t, pln.ID, "nil plan ID")
			assert.NotEmpty(t, pln.Name, "nil plan name")
			assert.NotEmpty(t, pln.Description, "nil plan description")
			assert.NotNil(t, pln.Free, "nil plan free indicator")

		}

	}

}
