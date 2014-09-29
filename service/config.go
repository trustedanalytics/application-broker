package service

import (
	"flag"
	"github.com/cloudfoundry-community/go-cfenv"
)

var ServiceConfig *Config = &Config{}

func init() {
	ServiceConfig.initialize()
}

type Config struct {
	Source           string
	DashboardRootURL string
	Debug            bool
	Env              *cfenv.App
}

func (c *Config) initialize() {
	flag.StringVar(&c.Source, "src", "spring-music", "Source App")
	flag.StringVar(&c.DashboardRootURL, "url", "http://domain.com", "Root URL")

	env, _ := cfenv.Current()
	c.Env = env
}
