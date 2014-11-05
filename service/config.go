package service

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cloudfoundry-community/types-cf"
)

const (
	// DefaultCatalogFilePath represents path to the default catalog file
	DefaultCatalogFilePath = "./catalog.json"
)

//
var Config = &ServiceConfig{}

func init() {
	Config.initialize()
}

// Dependency represents a dependancy
type Dependency struct {
	Name string
	Plan string
}

// ServiceConfig holds the service config
type ServiceConfig struct {
	APIEndpoint  string
	SkipSSLValid bool
	APIUser      string
	APIPassword  string
	AppSource    string
	DepString    string
	CFEnv        *cfenv.App
	CatalogPath  string
	Catalog      *cf.Catalog
	ServiceName  string
	Dependencies []*Dependency
}

func (c *ServiceConfig) initialize() {
	log.Println("initializing service config...")

	c.APIEndpoint = GetEnvVarAsString("CF_API", "")
	c.SkipSSLValid = GetEnvVarAsBool("CF_API_SKIP_SSL_VALID", false)
	c.APIUser = GetEnvVarAsString("CF_USER", "")
	c.APIPassword = GetEnvVarAsString("CF_PASS", "")
	c.AppSource = GetEnvVarAsString("CF_SRC", "")
	c.DepString = GetEnvVarAsString("CF_DEP", "")
	c.CatalogPath = GetEnvVarAsString("CF_CATALOG_PATH", "./catalog.json")

	cfEnv, err := cfenv.Current()
	if err == nil || cfEnv == nil {
		log.Printf("CF env vars: %v", err)
		cfEnv = &cfenv.App{}
		cfEnv.TempDir = os.TempDir()
	}
	c.CFEnv = cfEnv
	c.loadCatalogFromFile()
	c.ServiceName = c.Catalog.Services[0].Name
	c.parseServiceDependencies()
}

func (c *ServiceConfig) parseServiceDependencies() error {
	log.Println("getting service dependences")
	if len(c.DepString) < 1 {
		log.Println("nil dependencies")
		return nil
	}
	parts := strings.Split(c.DepString, ",")
	deps := make([]*Dependency, len(parts))
	for i, part := range parts {
		log.Printf("part[%d] %s", i, part)
		dep := strings.Split(part, "|")
		log.Printf("dep:%s plan:%s", dep[0], dep[1])
		deps[i] = &Dependency{Name: dep[0], Plan: dep[1]}
	}
	c.Dependencies = deps
	return nil
}

func (c *ServiceConfig) loadCatalogFromFile() {

	log.Printf("loading catalog from: %s", c.CatalogPath)

	if len(c.CatalogPath) < 1 {
		c.CatalogPath = path.Join(getServiceDir(), DefaultCatalogFilePath)
		log.Printf("no catalog path provided, using default: %s ...", c.CatalogPath)
	}

	if !pathExists(c.CatalogPath) {
		log.Printf("unable to find: %s", c.CatalogPath)
		return
	}

	log.Printf("reading: %s ...", c.CatalogPath)
	file, e := ioutil.ReadFile(c.CatalogPath)
	if e != nil {
		log.Printf("error on open: %v\n", e)
		return
	}

	catalog := &cf.Catalog{}
	json.Unmarshal(file, &catalog)

	c.Catalog = catalog
}
