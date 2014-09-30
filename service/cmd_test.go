package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleCommand(t *testing.T) {

	testVal := "test"
	cmd := newConsoleCommand("echo")
	cmd.setArgs(testVal).exec()

	assert.Nil(t, cmd.err, "command failed")
	assert.NotNil(t, cmd.output, "nil output")
	assert.Equal(t, cmd.output, "test", "wrong output")
}
