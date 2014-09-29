package service

import (
	"flag"
)

var ServiceConfig *Config = &Config{}

func init() {
	ServiceConfig.initialize()
}

type Config struct {
	Source string
	Debug  bool
}

func (o *Config) initialize() {
	flag.StringVar(&o.Source, "src", "spring-music", "Source application")
	flag.Parse()
}
