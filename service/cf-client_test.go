package service

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	TestOrg   = "demo"
	TestSpace = "dev"
)

func TestCFClient(t *testing.T) {

	flag.Set("api", "http://api.54.68.64.168.xip.io")
	flag.Set("cfu", os.Getenv("CF_USER"))
	flag.Set("cfp", os.Getenv("CF_PASS"))
	flag.Set("src", "/Users/markchma/Code/rabbitmq-cloudfoundry-samples/nodejs")
	flag.Set("app", "sinatra-cf-twitter")
	flag.Set("dep", "rabbitmq33|free,redis28|free")

	client := NewCFClient(ServiceConfig)
	assert.NotNil(t, client, "nil client")

	err := client.push(TestOrg, TestSpace)
	assert.Nil(t, err, "push failed")

}
