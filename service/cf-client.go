package service

import (
	"io/ioutil"
	"log"
)

// CFClient object
type CFClient struct {
	config *ServiceConfig
}

func NewCFClient(c *ServiceConfig) *CFClient {
	return &CFClient{
		config: c,
	}
}

func (c *CFClient) getSertviceInfo() error {
	log.Println("getting service info")

	// initialize
	cmd, err := c.initialize()
	if err != nil {
		log.Fatalf("err initializing command: %v", err)
		return err
	}

	// get app first
	cmd.setArgs("curl", "/v2/apps/:guid/summary").exec()
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}

	return nil

}

func (c *CFClient) initialize() (*consoleCommand, error) {
	log.Println("initializing...")

	// yep, this is a royal hack, should get this from the env somehow
	pushId := genRandomString(8)
	appDir, err := ioutil.TempDir(c.config.CFEnv.TempDir, pushId)
	if err != nil {
		log.Fatalf("err creating a temp dir: %v", err)
		return nil, err
	}

	// api
	cmd := newConsoleCommand("cf")

	// TODO: remove the skip API validation part once real cert deployed
	cmd.setArgs("api", c.config.ApiEndpoint, "--skip-ssl-validation").
		addEnv("CF_HOME", appDir).exec()
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd, cmd.err
	}

	// auth
	cmd.setArgs("auth", c.config.ApiUser, c.config.ApiPassword).exec()
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd, cmd.err
	}

	return cmd, nil
}

func (c *CFClient) deprovision(app, org, space string) error {
	log.Printf("deprovision app: %s/%s/%s", org, space, app)

	// initialize
	cmd, err := c.initialize()
	if err != nil {
		log.Fatalf("err initializing command: %v", err)
		return err
	}

	// target
	cmd.setArgs("target", "-o", org, "-s", space).exec()
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}

	// delete
	cmd.setArgs("d", app, "-f").exec()
	if cmd.err != nil {
		log.Printf("err cmd: %v", cmd)
		return cmd.err
	}

	for i, dep := range c.config.Dependencies {
		depName := dep.Name + "-" + app
		cmd.setArgs("delete-service", dep.Name, "-f").exec()
		if cmd.err != nil {
			log.Printf("err on dependency delete[%d]: %s - %v", i, depName, cmd)
		}
	}

	return nil
}

func (c *CFClient) provision(app, org, space string) error {
	log.Printf("provisioning app: %s/%s/%s", org, space, app)

	// initialize
	cmd, err := c.initialize()
	if err != nil {
		log.Fatalf("err initializing command: %v", err)
		return err
	}

	// target
	cmd.setArgs("target", "-o", org, "-s", space).exec()
	if cmd.err != nil {
		log.Fatalf("err cmd: %v", cmd)
		return cmd.err
	}

	// push
	cmd.setArgs("push", app, "-p", c.config.AppSource, "--no-start").exec()
	if cmd.err != nil {
		log.Printf("err cmd: %v", cmd)
		c.deprovision(app, org, space)
		return cmd.err
	}

	// TODO: Add cleanup of dependencies
	for i, dep := range c.config.Dependencies {
		depName := dep.Name + "-" + app
		cmd.setArgs("create-service", dep.Name, dep.Plan, depName).exec()
		if cmd.err != nil {
			log.Printf("err on dependency[%d]: %s - %v", i, depName, cmd)
			return cmd.err
		}

		// bind
		cmd.setArgs("bind-service", app, depName).exec()
		if cmd.err != nil {
			log.Printf("err on bind[%d]: %s > %s - %v", i, app, depName, cmd)
			return cmd.err
		}

		//TODO: check if we need to restage the app after binding
	}

	// start
	cmd.setArgs("start", app).exec()
	if cmd.err != nil {
		log.Printf("err cmd: %v", cmd)
		c.deprovision(app, org, space)
		return cmd.err
	}

	return nil
}
