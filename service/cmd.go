package service

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func newConsoleCommand(cmd string) *consoleCommand {
	return &consoleCommand{
		command: cmd,
		envs:    make(map[string]string),
	}
}

type consoleCommand struct {
	dir     string
	command string
	args    []string
	output  string
	err     error
	envs    map[string]string
}

func (c *consoleCommand) addEnv(key, val string) *consoleCommand {
	c.envs[key] = val
	return c
}

func (c *consoleCommand) setArgs(args ...string) *consoleCommand {
	c.args = args
	return c
}

func (c *consoleCommand) exec() {
	var wg sync.WaitGroup
	wg.Add(1)
	go c.execAsync(&wg)
	wg.Wait()
}

func (c *consoleCommand) execAsync(wg *sync.WaitGroup) {
	// don't log the command, could include passwords
	if wg == nil {
		return
	}
	if c == nil {
		wg.Done()
	}

	if len(c.dir) > 1 {
		os.Chdir(c.dir)
	}

	// set env vars right before executing this specific command
	if len(c.envs) > 0 {
		for key, value := range c.envs {
			setEnv(key, value)
		}
	}

	cmd := exec.Command(c.command, c.args...)
	out, err := cmd.Output()
	c.err = err
	// yep, this is a hack to pass simple tests
	// expecting users to encode results
	c.output = strings.Trim(string(out), "\n")
	log.Printf("cmd output: %s", c.output)
	wg.Done()
}
