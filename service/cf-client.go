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

	// yep, this is a royal hack, should get this from the env somehow
	raw := genRandomString(4)
	pushId := "-" + strings.ToUpper(strings.Trim(raw, "=="))
	appName := c.config.AppBaseName + pushId

	cmd := newCommand("cf", "target", "-o", org, "-s", space)
	exeCmd(cmd)
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}
	log.Printf("target output: %s", cmd.output)

	cmd = newCommand("cf", "push", appName, "-p", c.config.AppSource)
	exeCmd(cmd)
	if cmd.err != nil {
		log.Printf("err cmd: %v", cmd)
		// try to delete
		exeCmd(newCommand("cf", "d", appName))
		return cmd.err
	}
	log.Printf("push output: %s", cmd.output)

	deps, err := c.config.getDependencies()
	if err != nil {
		log.Printf("err cmd: %v", err)
		// try to delete the app
		exeCmd(newCommand("cf", "d", appName))
		return cmd.err
	}

	// TODO: Add cleanup of dependencies
	for i, dep := range deps {
		depName := dep.Name + pushId
		cmd = newCommand("cf", "create-service", dep.Name, dep.Plan, depName)
		exeCmd(cmd)
		if cmd.err != nil {
			log.Printf("err on dependency[%d]: %s - %v", i, depName, cmd)
			return cmd.err
		}
	}

	return nil
}
