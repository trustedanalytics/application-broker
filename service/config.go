package service

import (
	"encoding/json"
	"flag"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/intel-data/types-cf"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

const (
	DefaultCatalogFilePath = "./catalog.json"
)

var Config *ServiceConfig = &ServiceConfig{}

func init() {
	Config.initialize(flag.CommandLine)
}

type ServiceDependency struct {
	Name string
	Plan string
}

type ServiceConfig struct {
	ApiEndpoint  string
	ApiUser      string
	ApiPassword  string
	AppSource    string
	DepFlag      string
	CFEnv        *cfenv.App
	CatalogPath  string
	Catalog      *cf.Catalog
	Dependencies []*ServiceDependency
}

func (c *ServiceConfig) initialize(fs *flag.FlagSet) {
	log.Println("initializing service config...")

	fs.StringVar(&c.ApiEndpoint, "api", "", "Full URL to the API endpoint")
	fs.StringVar(&c.ApiUser, "cfu", "", "CF user (should be admin)")
	fs.StringVar(&c.ApiPassword, "cfp", "", "CF Password")
	fs.StringVar(&c.AppSource, "src", "", "Path to source of the app to provision")
	fs.StringVar(&c.DepFlag, "dep", "", "Service dependencies: (postgresql93|free,consul|free)")
	fs.StringVar(&c.CatalogPath, "cat", "", "Path to catalog file [./catalog.json]")
}

func (c *ServiceConfig) parse() {
	cfEnv, err := cfenv.Current()
	if err == nil || cfEnv == nil {
		log.Printf("failed to get CF env vars, probably running locally: %v", err)
		cfEnv = &cfenv.App{}
	}
	c.CFEnv = cfEnv
	c.loadCatalogFromFile()
	c.parseServiceDependencies()
}

func (c *ServiceConfig) parseServiceDependencies() error {
	log.Println("getting service dependences")
	if len(c.DepFlag) < 1 {
		log.Println("nil dependencies")
		return nil
	}
	parts := strings.Split(c.DepFlag, ",")
	deps := make([]*ServiceDependency, len(parts))
	for i, part := range parts {
		log.Printf("part[%d] %s", i, part)
		dep := strings.Split(part, "|")
		log.Printf("dep:%s plan:%s", dep[0], dep[1])
		deps[i] = &ServiceDependency{Name: dep[0], Plan: dep[1]}
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
