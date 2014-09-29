package service

import (
	"flag"
)

var ServiceConfig *Config = &Config{}

func init() {
	ServiceConfig.initialize()
}

type Config struct {
	Source           string
	DashboardRootURL string
	Debug            bool
}

func (o *Config) initialize() {
	flag.StringVar(&o.Source, "src", "spring-music", "Source application")
	flag.StringVar(&o.DashboardRootURL, "url", "https://somename.gotapaas.com", "Root of the app dashboard")
}
