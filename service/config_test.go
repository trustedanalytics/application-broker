package service

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

const (
	TestSourceApp = "bigapp"
)

func TestConfig(t *testing.T) {

	flag.Set("s", TestSourceApp)

	c := BrokerConfig

	assert.NotEmpty(t, c, "nil config")
	assert.Equal(t, c.Source, TestSourceApp, "Invalid source")

}
