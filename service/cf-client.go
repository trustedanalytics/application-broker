package service

import (
	"log"
	"strings"
)

// CFClient object
type CFClient struct {
	config *Config
}

func NewCFClient(c *Config) (*CFClient, error) {
	client := &CFClient{
		config: c,
	}
	err := client.initialize()
	return client, err
}

func (c *CFClient) initialize() error {
	log.Println("initializing client...")

	// target API
	// TODO: remove the skip API validation part once real cert deployed
	cmd := newCommand("cf", "api", c.config.ApiEndpoint, "--skip-ssl-validation")
	exeCmd(cmd)
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}
	log.Printf("api output: %s", cmd.output)

	cmd = newCommand("cf", "auth", c.config.ApiUser, c.config.ApiPassword)
	exeCmd(cmd)
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}
	log.Printf("auth output: %s", cmd.output)

	return nil
}

func (c *CFClient) push(org, space string) error {
	log.Println("pushing app...")

	cmd := newCommand("cf", "target", "-o", org, "-s", space)
	exeCmd(cmd)
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}
	log.Printf("target output: %s", cmd.output)

	cmd = newCommand("cd", c.config.AppSource)
	exeCmd(cmd)
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}
	log.Printf("cd output: %s", cmd.output)

	// yep, this is a royal hack, should get this from the env somehow
	raw := genRandomString(4)
	sufix := strings.ToUpper(strings.Trim(raw, "=="))
	cmd = newCommand("cf", "push", c.config.AppBaseName+sufix)
	exeCmd(cmd)
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}
	log.Printf("push output: %s", cmd.output)

	return nil
}
