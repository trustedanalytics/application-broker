package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	assert.NotEmpty(t, Config, "nil config")
	assert.NotEmpty(t, Config.CFEnv.Port, "nil port")
}
