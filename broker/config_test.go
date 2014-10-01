package broker

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestConfig(t *testing.T) {

	assert.NotEmpty(t, c, "nil config")
	assert.NotEmpty(t, c.Port, "nil port")
	assert.NotEmpty(t, c.Username, "nil username")
	assert.NotEmpty(t, c.Password, "nil password")

}
