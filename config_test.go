package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig(t *testing.T) {

	c := Config{}
	c.initialize()

	assert.NotEmpty(t, c, "nil config")
	assert.NotEmpty(t, c.Username, "nil username")
	assert.NotEmpty(t, c.Password, "nil password")
	assert.NotEmpty(t, c.Port, "nil port")
	assert.NotEmpty(t, c.Host, "nil host")

}
