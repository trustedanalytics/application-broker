package broker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig(t *testing.T) {

	c := BrokerConfig

	assert.NotEmpty(t, c, "nil config")
	assert.NotEmpty(t, c.Username, "nil username")
	assert.NotEmpty(t, c.Password, "nil password")
	assert.NotEmpty(t, c.Port, "nil port")
	assert.NotEmpty(t, c.Host, "nil host")

}
