// Copyright 2014, The go-cmd Authors. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package cmd

import (
	"os"
	"os/exec"
	"strings"
	"sync"
)

// New creates a new instance of the Command based
func New(cmd string) *Command {
	return &Command{
		Cmd:  cmd,
		Envs: make(map[string]string),
	}
}

// Command struct holds the command variables
type Command struct {
	Dir  string
	Cmd  string
	Args []string
	Out  string
	Err  error
	Envs map[string]string
}

// WithEnv adds environment variables that will be set before execution
func (c *Command) WithEnv(k, v string) *Command {
	c.Envs[k] = v
	return c
}

// WithArgs adds command arguments, helpful when you execute the same
// command multiple times with diff arguments
func (c *Command) WithArgs(args ...string) *Command {
	c.Args = args
	return c
}

// Exec creates WaitGroup group and executes the command asynchronously
func (c *Command) Exec() *Command {
	var wg sync.WaitGroup
	wg.Add(1)
	go c.ExecAsync(&wg)
	wg.Wait()
	return c
}

// ExecAsync executes the command asynchronously
func (c *Command) ExecAsync(wg *sync.WaitGroup) {

	if len(c.Dir) > 1 {
		os.Chdir(c.Dir)
	}

	// set env vars right before executing this specific command
	if len(c.Envs) > 0 {
		for k, v := range c.Envs {
			os.Setenv(k, v)
		}
	}

	cmd := exec.Command(c.Cmd, c.Args...)
	o, err := cmd.Output()
	c.Err = err
	// yep, this is a hack to pass simple tests
	// expecting users to encode results
	c.Out = strings.Trim(string(o), "\n")
	wg.Done()
}
