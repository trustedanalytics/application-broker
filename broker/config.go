package broker

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"log"
	"os"
)

// Config hold a global BrokerConfig isntance
var Config = &BrokerConfig{}

func init() {
	Config.initialize()
}

// BrokerConfig hold the broker configuration
type BrokerConfig struct {
	Username string
	Password string
	CFEnv    *cfenv.App
}

func (c *BrokerConfig) initialize() {
	log.Println("initializing broker config...")
	c.Username = os.Getenv("CF_USER")
	c.Password = os.Getenv("CF_PASS")

	cfEnv, err := cfenv.Current()
	if err == nil || cfEnv == nil {
		log.Printf("failed to get CF env vars, probably running locally: %v", err)
		cfEnv = &cfenv.App{}
		cfEnv.Port = 9999
	}
	c.CFEnv = cfEnv

}
