package service

import (
	"io/ioutil"
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

func (c *CFClient) push(org, space string) error {
	log.Println("pushing app...")

	// yep, this is a royal hack, should get this from the env somehow
	pushId := genRandomString(6)
	appName := c.config.AppBaseName + pushId
	appDir, err := ioutil.TempDir(c.config.CFEnv.TempDir, appName)
	if err != nil {
		log.Fatalf("err creating a temp dir: %v", err)
		return err
	}

	// api
	cmd := newConsoleCommand("cf")

	// TODO: remove the skip API validation part once real cert deployed
	cmd.setArgs("api", c.config.ApiEndpoint, "--skip-ssl-validation").
		addEnv("CF_HOME", appDir).exec()
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}

	// auth
	cmd.setArgs("auth", c.config.ApiUser, c.config.ApiPassword).exec()
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}

	// target
	cmd.setArgs("target", "-o", org, "-s", space).exec()
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}

	// push
	cmd.setArgs("push", appName, "-p", c.config.AppSource, "--no-start").exec()
	if cmd.err != nil {
		log.Printf("err cmd: %v", cmd)
		// try to delete
		cmd.setArgs("d", appName).exec()
		return cmd.err
	}

	// dependencies
	deps, err := c.config.getDependencies()
	if err != nil {
		log.Printf("err cmd: %v", err)
		// try to delete the app
		cmd.setArgs("d", appName).exec()
		return cmd.err
	}

	// TODO: Add cleanup of dependencies
	for i, dep := range deps {
		depName := dep.Name + pushId
		cmd.setArgs("create-service", dep.Name, dep.Plan, depName).exec()
		if cmd.err != nil {
			log.Printf("err on dependency[%d]: %s - %v", i, depName, cmd)
			return cmd.err
		}

		// bind
		cmd.setArgs("bind-service", appName, depName).exec()
		if cmd.err != nil {
			log.Printf("err on bind[%d]: %s > %s - %v", i, appName, depName, cmd)
			return cmd.err
		}
	}

	// start
	cmd.setArgs("start", appName).exec()
	if cmd.err != nil {
		log.Printf("err cmd: %v", cmd)
		// try to delete
		cmd.setArgs("d", appName).exec()
		return cmd.err
	}

	return nil
}
