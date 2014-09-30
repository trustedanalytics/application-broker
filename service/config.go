package service

import (
	"errors"
	"flag"
	"github.com/cloudfoundry-community/go-cfenv"
	"log"
	"strings"
)

var ServiceConfig *Config = &Config{}

func init() {
	ServiceConfig.initialize()
}

type ServiceDependency struct {
	Name string
	Plan string
}

type Config struct {
	AppSource    string
	Dependencies string
	CFEnv        *cfenv.App
}

func (c *Config) initialize() {
	flag.StringVar(&c.AppSource, "src", "", "Source of the app to push [./spring-music]")
	flag.StringVar(&c.Dependencies, "dep", "", "Service dependencies: [postgresql93|free,consul|free]")

	env, err := cfenv.Current()
	if err == nil || env == nil {
		log.Printf("failed to get CF env vars: %v", err)
		env = &cfenv.App{}
		env.Host = "127.0.0.1"
		env.Port = 9999
	}
	c.CFEnv = env
}

func (c *Config) getDependencies() ([]*ServiceDependency, error) {

	if len(c.Dependencies) < 1 {
		return nil, errors.New("nil dependencies")
	}

	parts := strings.Split(c.Dependencies, ",")
	deps := make([]*ServiceDependency, len(parts))

	for i, part := range strings.Split(c.Dependencies, ",") {
		log.Printf("part[%d] %s", i, part)
		dep := strings.Split(part, "|")
		log.Printf("dep:%s plan:%s", dep[0], dep[1])
		deps[i] = &ServiceDependency{Name: dep[0], Plan: dep[1]}
	}

	return deps, nil
}
