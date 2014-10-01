package broker

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"log"
	"os"
)

var Config *BrokerConfig = &BrokerConfig{}

func init() {
	Config.initialize()
}

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
		cfEnv.Port = 8888
	}
	c.CFEnv = cfEnv

}
