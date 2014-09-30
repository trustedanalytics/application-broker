package service

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCFClient(t *testing.T) {

	flag.Set("api", "http://api.54.68.64.168.xip.io")
	flag.Set("cfu", os.Getenv("CF_USER"))
	flag.Set("cfp", os.Getenv("CF_PASS"))
	flag.Set("src", "./spring-music")
	flag.Set("dep", "postgresql93|free,consul|free")

	client := NewCFClient(ServiceConfig)
	assert.NotNil(t, client, "nil client")

	err := client.initialize()
	assert.Nil(t, err, "client failed to initialize")

}
