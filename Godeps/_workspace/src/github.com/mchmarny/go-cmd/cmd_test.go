// Copyright 2014, The go-cmd Authors. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package cmd

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestSimpleCommand(t *testing.T) {

	testVal := "test"
	cmd := New("echo").
		WithEnv("MY_VAR", "something").
		WithArgs("-n").
		WithArgs(testVal).
		Exec()

	if cmd.Err != nil {
		log.Fatalf("command failed: %v", cmd)
	}

	log.Printf("command output: %s", cmd.Out)

	assert.Nil(t, cmd.Err, "command failed")
	assert.NotNil(t, cmd.Out, "nil output")
	assert.Equal(t, cmd.Out, "test", "wrong output")
}
