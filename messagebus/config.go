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
	"github.com/trustedanalytics/application-broker/env"
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

	if cfEnv == nil {
		c.url = env.GetEnvVarAsString("NATS_URL", "nats://localhost:4222")
		c.subject = env.GetEnvVarAsString("NATS_SERVICE_CREATION_SUBJECT", "service-creation")

		if len(c.url) == 0 || len(c.subject) == 0 {
			log.Debug("Unable to collect nats configuration from local envs")
			return false
		}
		return true
	}

	service, err := cfEnv.Services.WithName("nats-provider")
	if err != nil {
		log.Debug("Cannot locate nats service in CF environment")
		return false
	}
	c.url = service.Credentials["url"].(string)
	c.subject = service.Credentials["service-creation-subject"].(string)
	return true
}
