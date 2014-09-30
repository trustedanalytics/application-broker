package service

import (
	"log"
)

// CFClient object
type CFClient struct {
	config *Config
}

func NewCFClient(c *Config) *CFClient {
	return &CFClient{
		config: c,
	}
}

func (c *CFClient) initialize() error {
	log.Println("initializing client...")

	// target API
	// TODO: remove the skip API validation part once real cert deployed
	cmd := newCommand("cf", "api", c.config.ApiEndpoint, "--skip-ssl-validation")
	exeCmd(cmd)
	if cmd.err != nil {
		return cmd.err
	}
	log.Printf("api output: %s", cmd.output)

	cmd2 := newCommand("cf", "auth", c.config.ApiUser, c.config.ApiPassword)
	exeCmd(cmd2)
	if cmd2.err != nil {
		return cmd2.err
	}
	log.Printf("auth output: %s", cmd2.output)

	return nil
}
