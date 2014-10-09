package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCFServiceQuery(t *testing.T) {

	client := NewCFClient(Config)
	assert.NotNil(t, client, "nil client")

	resp, err := client.getServices()
	assert.Nil(t, err, "query failed")
	assert.NotNil(t, resp, "nil response")

	for _, r := range resp.Resources {

		srv, err2 := client.getService(r.Meta.GUID)
		assert.Nil(t, err2, "query resource failed")
		assert.NotNil(t, srv, "nil service")
		assert.NotEmpty(t, srv.Name, "nil service name")
		assert.NotEmpty(t, srv.SpaceGUID, "nil service space quid")

		sp, err3 := client.getSpace(srv.SpaceGUID)
		assert.Nil(t, err3, "query space failed")
		assert.NotNil(t, sp, "nil space")
		assert.NotEmpty(t, sp.Name, "nil space name")

		org, err4 := client.getOrg(sp.OrgGUID)
		assert.Nil(t, err4, "query space failed")
		assert.NotNil(t, org, "nil space")
		assert.NotEmpty(t, org.Name, "nil space name")

	}

}

func TestCFAppQuery(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping CF tests in short mode")
		return
	}

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
