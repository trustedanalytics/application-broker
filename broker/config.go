package broker

import (
	"flag"
)

var BrokerConfig *Config = &Config{}

func init() {
	BrokerConfig.initialize()
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Debug    bool
}

func (o *Config) initialize() {
	flag.StringVar(&o.Host, "h", "127.0.0.1", "Host")
	flag.IntVar(&o.Port, "p", 8888, "Port")
	flag.StringVar(&o.Username, "u", "operator", "User")
	flag.StringVar(&o.Password, "s", "secret", "Secret")
	flag.BoolVar(&o.Debug, "d", false, "Debug")
}
