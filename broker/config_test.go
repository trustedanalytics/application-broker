package broker

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func getTestBrokerConfig() *BrokerConfig {
	c := &BrokerConfig{}
	c.Host = "127.0.0.1"
	c.Port = "9999"
	c.Username = os.Getenv("CF_USER")
	c.Password = os.Getenv("CF_PASS")

	c.parse()

	// this is a total cudgel for testing
	Config = c

	return c
}

func TestConfig(t *testing.T) {

	c := getTestBrokerConfig()

	assert.NotEmpty(t, c, "nil config")
	assert.NotEmpty(t, c.Host, "nil hostname")
	assert.NotEmpty(t, c.Port, "nil port")
	assert.NotEmpty(t, c.Username, "nil username")
	assert.NotEmpty(t, c.Password, "nil password")

}
