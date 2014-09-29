package broker

import (
	"flag"
	"github.com/cloudfoundry-community/go-cfenv"
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
	Env      *cfenv.App
}

func (c *Config) initialize() {
	flag.StringVar(&c.Host, "h", "127.0.0.1", "Host")
	flag.IntVar(&c.Port, "p", 8888, "Port")
	flag.StringVar(&c.Username, "u", "operator", "User")
	flag.StringVar(&c.Password, "s", "secret", "Secret")
	flag.BoolVar(&c.Debug, "d", false, "Debug")

	env, _ := cfenv.Current()
	c.Env = env

}
