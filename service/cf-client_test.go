package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TestOrg   = "demo"
	TestSpace = "dev"
)

func TestCFClient(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping CF tests in short mode")
		return
	}

	client := NewCFClient(Config)
	assert.NotNil(t, client, "nil client")

	assert.NotNil(t, client.config.Catalog.Services, "nil services")
	assert.True(t, len(client.config.Catalog.Services) > 0, "services number")
	name := client.config.Catalog.Services[0].Name

	err := client.provision(name, TestOrg, TestSpace)
	assert.Nil(t, err, "provision failed")

	// regardless if the previous failed, cleanup
	err = client.deprovision(name, TestOrg, TestSpace)
	assert.Nil(t, err, "deprovision failed")

}

func TestCFAPIQuery(t *testing.T) {

	client := NewCFClient(Config)
	assert.NotNil(t, client, "nil client")

	resp, err := client.getApps()
	assert.Nil(t, err, "query failed")
	assert.NotNil(t, resp, "nil response")
	assert.NotEqual(t, resp.Count, 0, "response has 0 records")
	assert.Equal(t, resp.Pages, 1, "response paged")
	assert.Equal(t, len(resp.Resources), resp.Count, "record count and record number don't match")

	for _, r := range resp.Resources {

		app, err2 := client.getApp(r.Meta.GUID)
		assert.Nil(t, err2, "query resource failed")
		assert.NotNil(t, app, "nil app")
		assert.NotEmpty(t, app.Name, "nil app name")
		assert.NotEmpty(t, app.SpaceGUID, "nil app space quid")

		sp, err3 := client.getSpace(app.SpaceGUID)
		assert.Nil(t, err3, "query space failed")
		assert.NotNil(t, sp, "nil space")
		assert.NotEmpty(t, sp.Name, "nil space name")

		org, err4 := client.getOrg(sp.OrgGUID)
		assert.Nil(t, err4, "query space failed")
		assert.NotNil(t, org, "nil space")
		assert.NotEmpty(t, org.Name, "nil space name")

	}

}
