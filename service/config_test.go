package service

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestConfig(t *testing.T) {

	assert.NotEmpty(t, Config, "nil config")
	assert.NotNil(t, Config.CatalogPath, "nil catalog path")
	assert.NotNil(t, Config.Dependencies, "nil deps")
	assert.Equal(t, 2, len(Config.Dependencies), "incorrect number of deps")

	for i, dep := range Config.Dependencies {
		log.Printf("dep[%d]:%s (%s)", i, dep.Name, dep.Plan)
		assert.NotNil(t, dep.Name, "nil name")
		assert.NotNil(t, dep.Plan, "nil plan")
	}

	assert.NotNil(t, Config.CFEnv, "nil CFEnv")

}
