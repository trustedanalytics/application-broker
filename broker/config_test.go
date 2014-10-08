package broker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig(t *testing.T) {

	assert.NotEmpty(t, Config, "nil config")
	assert.NotEmpty(t, Config.CFEnv.Port, "nil port")
}
