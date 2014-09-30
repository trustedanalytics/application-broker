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
	flag.Set("src", "/Users/markchma/Code/spring-hello-env")
	flag.Set("app", "spring-hello-env-")
	flag.Set("dep", "postgresql93|free,consul|free")

	client, err := NewCFClient(ServiceConfig)
	assert.NotNil(t, client, "nil client")
	assert.Nil(t, err, "client failed to initialize")

	err2 := client.push(TestOrg, TestSpace)
	assert.Nil(t, err2, "push failed")

}
