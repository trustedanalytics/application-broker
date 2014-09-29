package service

import (
	"flag"
)

var ServiceConfig *Config = &Config{}

func init() {
	ServiceConfig.initialize(flag.CommandLine)
}

type Config struct {
	Source           string
	DashboardRootURL string
	Debug            bool
}

func (o *Config) initialize(f *flag.FlagSet) {
	f.StringVar(&o.Source, "src", "spring-music", "Source application")
	f.StringVar(&o.DashboardRootURL, "url", "https://somename.gotapaas.com", "Root of the app dashboard")
}
