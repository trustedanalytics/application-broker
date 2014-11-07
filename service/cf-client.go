package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/intel-data/app-launching-service-broker/utils"
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

func (c *CFClient) initialize() (*utils.CommandLogger, error) {
	log.Println("initializing...")

	// yep, this is a royal hack, should get this from the env somehow
	pushID := genRandomString(8)
	appDir, err := ioutil.TempDir(c.config.CFEnv.TempDir, pushID)
	if err != nil {
		log.Fatalf("err creating a temp dir: %v", err)
		return nil, err
	}

	// api
	cf := utils.NewCommandLogger("cf")

	// TODO: remove the skip API validation part once real cert deployed
	cf.WithArgs("api", c.config.APIEndpoint, "--skip-ssl-validation").
		WithEnv("CF_HOME", appDir).Exec()
	if cf.Err() != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf, cf.Err()
	}

	// auth
	cf.WithArgs("auth", c.config.APIUser, c.config.APIPassword).Exec()
	if cf.Err() != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf, cf.Err()
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
	if cf.Err() != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf.Err()
	}

	// push
	cf.WithArgs("push", ctx.InstanceName, "-p", c.config.AppSource, "--no-start").Exec()
	if cf.Err() != nil {
		log.Printf("err cmd: %v", cf)
		c.deprovision(ctx)
		return cf.Err()
	}

	// TODO: Add cleanup of dependencies
	for i, dep := range c.config.Dependencies {
		depName := dep.Name + "-" + ctx.InstanceName
		cf.WithArgs("create-service", dep.Name, dep.Plan, depName).Exec()
		if cf.Err() != nil {
			log.Printf("err on dependency[%d]: %s - %v", i, depName, cf)
			return cf.Err()
		}

		// bind
		cf.WithArgs("bind-service", ctx.InstanceName, depName).Exec()
		if cf.Err() != nil {
			log.Printf("err on bind[%d]: %s > %s - %v", i, ctx.InstanceName, depName, cf)
			return cf.Err()
		}

		//TODO: check if we need to restage the app after binding
	}

	// start
	cf.WithArgs("start", ctx.InstanceName).Exec()
	if cf.Err() != nil {
		log.Printf("err cmd: %v", cf)
		c.deprovision(ctx)
		return cf.Err()
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
	if cf.Err() != nil {
		log.Fatalf("err cmd: %v", cf)
		return cf.Err()
	}

	// delete
	cf.WithArgs("delete", ctx.InstanceName, "-f").Exec()
	if cf.Err() != nil {
		log.Printf("err cmd: %v", cf)
		return cf.Err()
	}

	// TODO: Does the service have to unbined first
	//       or deleting app will take care of it
	for i, dep := range c.config.Dependencies {
		depName := dep.Name + "-" + ctx.InstanceName
		cf.WithArgs("delete-service", dep.Name, "-f").Exec()
		if cf.Err() != nil {
			log.Printf("err on dependency delete[%d]: %s - %v", i, depName, cf)
		}
	}

	return nil
}

func (c *CFClient) queryAPI(query string) (string, error) {
	log.Printf("running query: %s", query)
	cf, err := c.initialize()
	if err != nil {
		log.Fatalf("err initializing command: %v", err)
		return "", err
	}
	cf.WithArgs("curl", query).Exec()
	return cf.Out(), cf.Err()
}

func (c *CFClient) getContextFromSpaceOrg(instanceID, spaceGUID, orgGUID string) (*CFServiceContext, error) {
	log.Printf("getting service context for ID %s in org %s space %s", instanceID, orgGUID, spaceGUID)

	t := NewCFServiceContext(instanceID)

	space, err := c.getSpace(spaceGUID)
	if err != nil {
		log.Printf("error getting space: %v", err)
		return nil, err
	}
	t.SpaceName = space.Name

	org, err := c.getOrg(orgGUID)
	if err != nil {
		log.Printf("error getting org: %v", err)
		return nil, err
	}
	t.OrgName = org.Name

	return t, nil

}

func (c *CFClient) getContextFromServiceInstanceID(instanceID string) (*CFServiceContext, error) {
	log.Printf("getting service context for: %s", instanceID)

	t := &CFServiceContext{}
	t.InstanceID = instanceID

	srv, err := c.getService(instanceID)
	if err != nil {
		log.Printf("error getting service: %v", err)
		return nil, err
	}
	t.InstanceName = srv.Name

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

func (c *CFClient) getService(instanceID string) (*cfApp, error) {
	log.Printf("getting service info for: %s", instanceID)
	query := fmt.Sprintf("/v2/service_instances/%s", instanceID)
	resp, err := c.queryAPI(query)
	if err != nil {
		return nil, errors.New("query error")
	}

	// cf-client.go:150: running query: /v2/service_instances/26576e51...
	// cf-client.go:26: initializing...
	// {
	//    "code": 60004,
	//    "description": "The service instance could not be found: 26576e51-8a47-46e3-bd6e-5908287e9935",
	//    "error_code": "CF-ServiceInstanceNotFound"
	// }
	//
	// TODO: map results to a CFError struct to see if an error was returned.
	// FIXME: looks like service instance object doesn't exist when "cf create-service" called
	// TODO: perhaps a background worker to rename service instances later?

	t := &cfAppResource{}
	log.Println(string(resp))
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("service output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getOrg(orgID string) (*cfApp, error) {
	log.Printf("getting org info for: %s", orgID)
	query := fmt.Sprintf("/v2/organizations/%s", orgID)
	resp, err := c.queryAPI(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	log.Println(string(resp))
	t := &cfAppResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("org output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getSpace(spaceID string) (*cfSpace, error) {
	log.Printf("getting space info for: %s", spaceID)
	query := fmt.Sprintf("/v2/spaces/%s", spaceID)
	resp, err := c.queryAPI(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	log.Println(string(resp))
	t := &cfSpaceResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("space output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getApp(appID string) (*cfApp, error) {
	log.Printf("getting app info for: %s", appID)
	query := fmt.Sprintf("/v2/apps/%s", appID)
	resp, err := c.queryAPI(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	log.Println(string(resp))
	t := &cfAppResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("app output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getBinding(bindingID string) (*CFBindingResponse, error) {
	log.Printf("getting service binding for: %s", bindingID)
	query := fmt.Sprintf("/v2/service_bindings/%s", bindingID)
	resp, err := c.queryAPI(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	log.Println(string(resp))
	t := &cfBindingResource{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("service binding output: %v", t)
	t.Entity.GUID = t.Meta.GUID
	return &t.Entity, nil
}

func (c *CFClient) getApps() (*cfAppsResponse, error) {
	log.Println("getting apps...")
	query := "/v2/apps?results-per-page=100"
	resp, err := c.queryAPI(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	log.Println(string(resp))
	t := &cfAppsResponse{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("apps output: %v", t)
	return t, nil
}

func (c *CFClient) getServices() (*cfAppsResponse, error) {
	log.Println("getting services...")
	query := "/v2/service_instances?results-per-page=100"
	resp, err := c.queryAPI(query)
	if err != nil {
		return nil, errors.New("query error")
	}
	log.Println(string(resp))
	t := &cfAppsResponse{}
	err2 := json.Unmarshal([]byte(resp), &t)
	if err2 != nil {
		log.Fatalf("err unmarshaling: %v - %v", err2, resp)
		return nil, errors.New("invalid JSON")
	}
	log.Printf("services output: %v", t)
	return t, nil
}
