package service

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TestSourceApp = "bigapp"
)

func TestConfig(t *testing.T) {

	flag.Set("src", TestSourceApp)

	c := ServiceConfig

	assert.NotEmpty(t, c, "nil config")
	assert.Equal(t, c.Source, TestSourceApp, "Invalid source")

}
