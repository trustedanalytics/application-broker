package broker

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

const (
	TestHostName = "bigserver"
	TestPort     = 7777
	TestUsername = "test-username"
	TestPassword = "test-password"
)

func TestConfig(t *testing.T) {

	flag.Set("h", TestHostName)
	flag.Set("p", strconv.Itoa(TestPort))
	flag.Set("u", TestUsername)
	flag.Set("s", TestPassword)

	c := BrokerConfig

	assert.NotEmpty(t, c, "nil config")
	assert.Equal(t, c.Host, TestHostName, "Invalid hostname")
	assert.Equal(t, c.Port, TestPort, "invalid port")
	assert.Equal(t, c.Username, TestUsername, "invalid username")
	assert.Equal(t, c.Password, TestPassword, "invalid password")

}
