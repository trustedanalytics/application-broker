/**
 * Copyright (c) 2015 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package messagebus handles emitting events to buses like NATS
package messagebus

import (
	log "github.com/cihub/seelog"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/trustedanalytics/application-broker/misc"
)

// Config holds values needed to set-up connection with MessageBus
type Config struct {
	url     string
	subject string
}

// TryInitialize initialize connection with message bus.
// It may fail when bad url was given for example.
// It will return true on success, false otherwise. */
func (c *Config) TryInitialize(cfEnv *cfenv.App) bool {
	log.Info("initializing nats config...")

	service := c.getNatsService(cfEnv)
	c.url = c.getUri(service)
	c.subject = c.getSubject(service)

	log.Infof("Fetched nats config: [%v] [%v]", c.url, c.subject)

	return len(c.url) > 0 && len(c.subject) > 0
}

func (c *Config) getUri(service *cfenv.Service) string {
	var url string
	if service != nil {
		url = service.Credentials["url"].(string)
	}
	if len(url) == 0 {
		url = misc.GetEnvVarAsString("NATS_URL", "nats://localhost:4222")
	}
	return url
}

func (c *Config) getSubject(service *cfenv.Service) string {
	var subject string
	if service != nil {
		subject = service.Credentials["service-creation-subject"].(string)
	}
	if len(subject) == 0 {
		subject = misc.GetEnvVarAsString("NATS_SERVICE_CREATION_SUBJECT", "service-creation")
	}
	return subject
}

func (c *Config) getNatsService(cfEnv *cfenv.App) *cfenv.Service {
	if cfEnv == nil {
		return nil
	}

	service, err := cfEnv.Services.WithName("nats-provider")
	if err != nil {
		log.Warn("Cannot locate nats service in CF environment")
	}
	return service
}
