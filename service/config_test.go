package service

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

const (
	TestAppSource    = "bigapp"
	TestDependancies = "postgresql93|free,consul|free"
)

func TestConfig(t *testing.T) {

	flag.Set("src", TestAppSource)
	flag.Set("dep", TestDependancies)

	c := ServiceConfig

	assert.NotEmpty(t, c, "nil config")
	assert.Equal(t, c.AppSource, TestAppSource, "Invalid source")
	assert.Equal(t, c.Dependencies, TestDependancies, "Invalid dependencies")

	deps, err := c.getDependencies()
	assert.Nil(t, err, err)
	assert.NotNil(t, deps, "nil deps")
	assert.Equal(t, 2, len(deps), "incorrect number of deps")

	for i, dep := range deps {
		log.Printf("dep[%d]:%s (%s)", i, dep.Name, dep.Plan)
		assert.NotNil(t, dep.Name, "nil name")
		assert.NotNil(t, dep.Plan, "nil plan")
	}

	assert.NotNil(t, c.CFEnv, "nil CFEnv")

}
