package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TestServiceId = "291E642F-E786-49EF-B6FC-F8AF96A36A37"
)

func TestGetServiceDashboard(t *testing.T) {

	p := &SimpleServiceProvider{}
	p.Initialize()

	srv, err := p.GetServiceDashboard(TestServiceId)

	assert.Nil(t, err, err)
	assert.NotNil(t, srv, "nil service")
	assert.NotEmpty(t, srv.DashboardURL, "missing dashboard element")

}