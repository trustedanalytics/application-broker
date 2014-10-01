package broker

import (
	"flag"
	"github.com/cloudfoundry-community/go-cfenv"
	"log"
)

var Config *BrokerConfig = &BrokerConfig{}

func init() {
	Config.initialize(flag.CommandLine)
}

type BrokerConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	CFEnv    *cfenv.App
}

func (c *BrokerConfig) initialize(fs *flag.FlagSet) {
	log.Println("initializing broker config...")
	fs.StringVar(&c.Host, "h", "127.0.0.1", "Host")
	fs.IntVar(&c.Port, "p", 8888, "Port")
	fs.StringVar(&c.Username, "u", "operator", "User")
	fs.StringVar(&c.Password, "s", "secret", "Secret")
}

func (c *BrokerConfig) parse() {
	cfEnv, err := cfenv.Current()
	if err == nil || cfEnv == nil {
		log.Printf("failed to get CF env vars, probably running locally: %v", err)
		cfEnv = &cfenv.App{}
	}
	c.CFEnv = cfEnv
}
