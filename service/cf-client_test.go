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
