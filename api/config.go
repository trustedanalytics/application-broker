package api

import (
	"log"
	"os"
)

var Config = &APIConfig{}

func init() {
	Config.initialize()
}

// APIConfig hold the broker configuration
type APIConfig struct {
	ApiURL string
	Debug  bool
}

func (c *APIConfig) initialize() {
	log.Println("initializing broker config...")

	c.ApiURL = os.Getenv("API_URL")

	c.validate()
}

func (c *APIConfig) validate() {
	missingEnvVars := []string{}

	if c.ApiURL == "" {
		missingEnvVars = append(missingEnvVars, "API_URL")
	}
	if len(missingEnvVars) > 0 {
		log.Println("Missing environment variable configuration:")
		for _, envVar := range missingEnvVars {
			log.Printf("* %s", envVar)
		}
		os.Exit(1)
	}
}
