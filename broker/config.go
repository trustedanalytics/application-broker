package broker

import (
	"log"
	"os"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/intel-data/app-launching-service-broker/service"
)

// Config hold a global BrokerConfig isntance
var Config = &BrokerConfig{}

func init() {
	Config.initialize()
}

// BrokerConfig hold the broker configuration
type BrokerConfig struct {
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	Debug        bool
	CFEnv        *cfenv.App
}

func (c *BrokerConfig) initialize() {
	log.Println("initializing broker config...")
	c.Username = os.Getenv("CF_USER")
	c.Password = os.Getenv("CF_PASS")
	c.Debug = os.Getenv("CF_DEBUG") == "true"

	c.ClientID = os.Getenv("CLIENT_ID")
	c.ClientSecret = os.Getenv("CLIENT_SECRET")
	c.RedirectURL = os.Getenv("REDIRECT_URL")
	c.AuthURL = os.Getenv("AUTH_URL")
	c.TokenURL = os.Getenv("TOKEN_URL")

	cfEnv, err := cfenv.Current()
	if err != nil || cfEnv == nil {
		log.Printf("failed to get CF env vars, probably running locally: %v", err)
		cfEnv = &cfenv.App{}
		cfEnv.Port = service.GetEnvVarAsInt("PORT", 9999)
		cfEnv.Host = "0.0.0.0"
		cfEnv.TempDir = os.TempDir()
	}
	c.CFEnv = cfEnv

}
