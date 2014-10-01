package service

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func getTestServiceConfig() *ServiceConfig {

	c := &ServiceConfig{}
	c.ApiEndpoint = "http://api.54.68.64.168.xip.io"
	c.ApiPassword = os.Getenv("CF_PASS")
	c.ApiUser = os.Getenv("CF_USER")
	c.AppSource = "/Users/markchma/Code/rabbitmq-cloudfoundry-samples/nodejs"
	c.CatalogPath = "../catalog.json"
	c.DepFlag = "rabbitmq33|free,redis28|free"

	c.parse()

	// this is a total cudgel for testing
	Config = c

	return c

}

func TestConfig(t *testing.T) {

	c := getTestServiceConfig()

	assert.NotEmpty(t, c, "nil config")
	assert.NotNil(t, c.CatalogPath, "nil catalog path")
	assert.NotNil(t, c.Dependencies, "nil deps")
	assert.Equal(t, 2, len(c.Dependencies), "incorrect number of deps")

	for i, dep := range c.Dependencies {
		log.Printf("dep[%d]:%s (%s)", i, dep.Name, dep.Plan)
		assert.NotNil(t, dep.Name, "nil name")
		assert.NotNil(t, dep.Plan, "nil plan")
	}

	assert.NotNil(t, c.CFEnv, "nil CFEnv")

}
