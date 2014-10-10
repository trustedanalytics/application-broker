package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mchmarny/go-cmd"
	"io/ioutil"
	"log"
)

// CFClient object
type CFClient struct {
	config *ServiceConfig
}

// NewCFClient creates a new isntance of CFClient
func NewCFClient(c *ServiceConfig) *CFClient {
	return &CFClient{
		config: c,
	}
}

func (c *CFClient) initialize() (*cmd.Command, error) {
	log.Println("initializing...")

	// yep, this is a royal hack, should get this from the env somehow
	pushID := genRandomString(8)
	appDir, err := ioutil.TempDir(c.config.CFEnv.TempDir, pushID)
	if err != nil {
		log.Fatalf("err creating a temp dir: %v", err)
		return nil, err
	}

	// api
	cf := cmd.New("cf")

	// TODO: remove the skip API validation part once real cert deployed
	cf.WithArgs("api", c.config.APIEndpoint, "--skip-ssl-validation").
		WithEnv("CF_HOME", appDir).Exec()
	if cf.Err != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf, cf.Err
	}

	// auth
	cf.WithArgs("auth", c.config.APIUser, c.config.APIPassword).Exec()
	if cf.Err != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf, cf.Err
	}

	return cf, nil
}

func (c *CFClient) provision(ctx *CFServiceContext) error {
	log.Printf("provisioning service: %v", ctx)

	// initialize
	cf, err := c.initialize()
	if err != nil {
		log.Fatalf("err initializing command: %v", err)
		return err
	}

	// target
	cf.WithArgs("target", "-o", ctx.OrgName, "-s", ctx.SpaceName).Exec()
	if cf.Err != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf.Err
	}

	// push
	cf.WithArgs("push", ctx.ServiceName, "-p", c.config.AppSource, "--no-start").Exec()
	if cf.Err != nil {
		log.Printf("err cmd: %v", cf)
		c.deprovision(ctx)
		return cf.Err
	}

	// TODO: Add cleanup of dependencies
	for i, dep := range c.config.Dependencies {
		depName := dep.Name + "-" + ctx.ServiceName
		cf.WithArgs("create-service", dep.Name, dep.Plan, depName).Exec()
		if cf.Err != nil {
			log.Printf("err on dependency[%d]: %s - %v", i, depName, cf)
			return cf.Err
		}

		// bind
		cf.WithArgs("bind-service", ctx.ServiceName, depName).Exec()
		if cf.Err != nil {
			log.Printf("err on bind[%d]: %s > %s - %v", i, ctx.ServiceName, depName, cf)
			return cf.Err
		}

		//TODO: check if we need to restage the app after binding
	}

	// start
	cf.WithArgs("start", ctx.ServiceName).Exec()
	if cf.Err != nil {
		log.Printf("err cmd: %v", cf)
		c.deprovision(ctx)
		return cf.Err
	}

	return nil
}

func (c *CFClient) deprovision(ctx *CFServiceContext) error {
	log.Printf("deprovision service: %v", ctx)

	// initialize
	cf, err := c.initialize()
	if err != nil {
		log.Fatalf("err initializing command: %v", err)
		return err
	}

	// target
	cf.WithArgs("target", "-o", ctx.OrgName, "-s", ctx.SpaceName).Exec()
	if cf.Err != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf.Err
	}

	// delete
	cf.WithArgs("delete", ctx.ServiceName, "-f").Exec()
	if cf.Err != nil {
		log.Printf("err cmd: %v", cf)
		return cf.Err
	}

	// TODO: Does the service have to unbined first
	//       or deleting app will take care of it
	for i, dep := range c.config.Dependencies {
		depName := dep.Name + "-" + ctx.ServiceName
		cf.WithArgs("delete-service", dep.Name, "-f").Exec()
		if cf.Err != nil {
			log.Printf("err on dependency delete[%d]: %s - %v", i, depName, cf)
		}
	}

	return nil
}

func (c *CFClient) binding(ctx *CFServiceContext, app *CFApp, on bool) error {
	log.Printf("binding service: %v to %v", ctx, app)

	// initialize
	cf, err := c.initialize()
	if err != nil {
		log.Fatalf("err initializing command: %v", err)
		return err
	}

	// target
	cf.WithArgs("target", "-o", ctx.OrgName, "-s", ctx.SpaceName).Exec()
	if cf.Err != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf.Err
	}

	bindCmd := "unbind-service"
	if on {
		bindCmd = "bind-service"
	}

	// bind
	cf.WithArgs(bindCmd, app.Name, ctx.ServiceName).Exec()
	if cf.Err != nil {
		log.Printf("err cmd: %v", cf)
		return cf.Err
	}

	return nil
}

func (c *CFClient) runQuery(query string) (string, error) {
	log.Printf("running query: %s", query)
	cf, err := c.initialize()
	if err != nil {
		log.Fatalf("err initializing command: %v", err)
		return "", err
	}
	cf.WithArgs("curl", query).Exec()
	return cf.Out, cf.Err
}

func (c *CFClient) getContext(instanceID string) (*CFServiceContext, error) {
	log.Printf("getting service context for: %s", instanceID)

	t := &CFServiceContext{}
	t.InstanceID = instanceID

	srv, err := c.getService(instanceID)
	if err != nil {
		log.Printf("error getting service: %v", err)
		return nil, err
	}
	t.ServiceName = srv.Name
	t.ServiceURI = srv.URI

	space, err := c.getSpace(srv.SpaceGUID)
	if err != nil {
		log.Printf("error getting space: %v", err)
		return nil, err
	}
	t.SpaceName = space.Name

	org, err := c.getOrg(space.OrgGUID)
	if err != nil {
		log.Printf("error getting org: %v", err)
		return nil, err
	}
	t.OrgName = org.Name

	return t, nil

}

func (c *CFClient) getService(instanceID string) (*CFApp, error) {
	log.Printf("getting service info for: %s", instanceID)
	query := fmt.Sprintf("/v2/service_instances/%s", instanceID)
	resp, err := c.runQuery(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	t := &CFAppResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("service output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getOrg(orgID string) (*CFApp, error) {
	log.Printf("getting org info for: %s", orgID)
	query := fmt.Sprintf("/v2/organizations/%s", orgID)
	resp, err := c.runQuery(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	t := &CFAppResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("org output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getSpace(spaceID string) (*CFSpace, error) {
	log.Printf("getting space info for: %s", spaceID)
	query := fmt.Sprintf("/v2/spaces/%s", spaceID)
	resp, err := c.runQuery(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	t := &CFSpaceResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("space output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getApp(appID string) (*CFApp, error) {
	log.Printf("getting app info for: %s", appID)
	query := fmt.Sprintf("/v2/apps/%s", appID)
	resp, err := c.runQuery(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	t := &CFAppResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("app output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getBinding(bindingID string) (*CFBinding, error) {
	log.Printf("getting service binding for: %s", bindingID)
	query := fmt.Sprintf("/v2/service_bindings/%s", bindingID)
	resp, err := c.runQuery(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	t := &CFBindingResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("service binding output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getApps() (*CFAppsResponce, error) {
	log.Println("getting apps...")
	query := "/v2/apps?results-per-page=100"
	resp, err := c.runQuery(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	t := &CFAppsResponce{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("apps output: %v", t)
	return t, nil
}

func (c *CFClient) getServices() (*CFAppsResponce, error) {
	log.Println("getting services...")
	query := "/v2/service_instances?results-per-page=100"
	resp, err := c.runQuery(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	t := &CFAppsResponce{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("services output: %v", t)
	return t, nil
}
