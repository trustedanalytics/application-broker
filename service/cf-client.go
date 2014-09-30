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
	cmd1 := newCommand("cf", "api", c.config.ApiEndpoint, "--skip-ssl-validation")
	exeCmd(cmd1)
	if cmd1.err != nil {
		log.Fatalf("err cmd: %v", cmd1)
		return cmd1.err
	}
	log.Printf("api output: %s", cmd1.output)

	cmd2 := newCommand("cf", "auth", c.config.ApiUser, c.config.ApiPassword)
	exeCmd(cmd2)
	if cmd2.err != nil {
		log.Fatalf("err cmd: %v", cmd2)
		return cmd2.err
	}
	log.Printf("auth output: %s", cmd2.output)

	cmd3 := newCommand("cf", "target", "-o", c.config.ApiOrg, "-s", c.config.ApiSpace)
	exeCmd(cmd3)
	if cmd3.err != nil {
		log.Fatalf("err cmd: %v", cmd3)
		return cmd3.err
	}
	log.Printf("target output: %s", cmd3.output)

	return nil
}
