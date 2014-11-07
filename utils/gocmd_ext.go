package utils

import (
	"log"
	"sync"

	"github.com/mchmarny/go-cmd"
)

// CommandLogger wraps cmd.Command
type CommandLogger struct {
	Command *cmd.Command
}

// NewCommandLogger creates a wrapper around a new CommandLogger
func NewCommandLogger(command string) *CommandLogger {
	return &CommandLogger{
		Command: cmd.New(command),
	}
}

// WithEnv adds environment variables that will be set before execution
func (c *CommandLogger) WithEnv(k, v string) *CommandLogger {
	log.Printf("%s.WithEnv(%s, %s)", c.Command.Cmd, k, v)
	c.Command.WithEnv(k, v)
	return c
}

// WithArgs adds command arguments, helpful when you execute the same
// command multiple times with diff arguments
func (c *CommandLogger) WithArgs(args ...string) *CommandLogger {
	log.Printf("%s.WithArgs(%v)", c.Command.Cmd, args)
	c.Command.WithArgs(args...)
	return c
}

// Exec creates WaitGroup group and executes the command asynchronously
func (c *CommandLogger) Exec() *CommandLogger {
	log.Printf("%s.Exec()", c.Command.Cmd)
	c.Command.Exec()
	return c
}

// ExecAsync executes the command asynchronously
func (c *CommandLogger) ExecAsync(wg *sync.WaitGroup) {
	log.Printf("%s.ExecAsync(%v)", c.Command.Cmd, wg)
	c.Command.ExecAsync(wg)
}

// Out returns the output from last executed command
func (c *CommandLogger) Out() string {
	return c.Command.Out
}

// Err returns the error from last executed command
func (c *CommandLogger) Err() error {
	return c.Command.Err
}
