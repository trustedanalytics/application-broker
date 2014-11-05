package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	NumberOfTestLoops = 2
)

func TestCFContextQuery(t *testing.T) {

	max := GetEnvVarAsInt("TEST_LOOP", NumberOfTestLoops)

	client := NewCFClient(Config)
	assert.NotNil(t, client, "nil client")

	resp, err := client.getServices()
	assert.Nil(t, err, "query failed")
	assert.NotNil(t, resp, "nil response")

	for i, r := range resp.Resources {

		ctx, err5 := client.getContext(r.Meta.GUID)
		assert.Nil(t, err5, "context query failed")
		assert.NotNil(t, ctx, "nil context")
		assert.NotEmpty(t, ctx.OrgName, "nil context org name")
		assert.NotEmpty(t, ctx.SpaceName, "nil context space name")
		assert.NotEmpty(t, ctx.SpaceName, "nil context service name")

		if i >= max {
			break
		}

	}

}

func TestCFServiceQuery(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping CF tests in short mode")
		return
	}

	max := GetEnvVarAsInt("TEST_LOOP", NumberOfTestLoops)

	client := NewCFClient(Config)
	assert.NotNil(t, client, "nil client")

	resp, err := client.getServices()
	assert.Nil(t, err, "query failed")
	assert.NotNil(t, resp, "nil response")

	for i, r := range resp.Resources {

		srv, err2 := client.getService(r.Meta.GUID)
		assert.Nil(t, err2, "query resource failed")
		assert.NotNil(t, srv, "nil service")
		assert.NotEmpty(t, srv.Name, "nil service name")
		assert.NotEmpty(t, srv.SpaceGUID, "nil service space quid")
		assert.NotEmpty(t, srv.URI, "nil service dashbaord")

		sp, err3 := client.getSpace(srv.SpaceGUID)
		assert.Nil(t, err3, "query space failed")
		assert.NotNil(t, sp, "nil space")
		assert.NotEmpty(t, sp.Name, "nil space name")

		org, err4 := client.getOrg(sp.OrgGUID)
		assert.Nil(t, err4, "query space failed")
		assert.NotNil(t, org, "nil space")
		assert.NotEmpty(t, org.Name, "nil space name")

		ctx, err5 := client.getContext(r.Meta.GUID)
		assert.Nil(t, err5, "context query failed")
		assert.NotNil(t, ctx, "nil context")
		assert.NotEmpty(t, ctx.OrgName, "nil context org name")
		assert.NotEmpty(t, ctx.SpaceName, "nil context space name")
		assert.NotEmpty(t, ctx.SpaceName, "nil context service name")

		assert.Equal(t, ctx.InstanceName, srv.Name, "context and service names should be the same")
		assert.Equal(t, ctx.ServiceURI, srv.URI, "context and service urls should be the same")
		assert.Equal(t, ctx.SpaceName, sp.Name, "context and space names should be the same")
		assert.Equal(t, ctx.OrgName, org.Name, "context and org names should be the same")

		if i >= max {
			break
		}

	}

}

func TestcfAppQuery(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping CF tests in short mode")
		return
	}

	max := GetEnvVarAsInt("TEST_LOOP", NumberOfTestLoops)

	client := NewCFClient(Config)
	assert.NotNil(t, client, "nil client")

	resp, err := client.getApps()
	assert.Nil(t, err, "query failed")
	assert.NotNil(t, resp, "nil response")
	assert.NotEqual(t, resp.Count, 0, "response has 0 records")
	assert.Equal(t, resp.Pages, 1, "response paged")
	assert.Equal(t, len(resp.Resources), resp.Count, "record count and record number don't match")

	for i, r := range resp.Resources {

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

		if i >= max {
			break
		}

	}

}
