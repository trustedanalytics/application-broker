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
package nats

import (
	"log"

	"github.com/cloudfoundry-community/go-cfenv"
	"os"
)

var Config = &NatsConfig{}

func init() {
	Config.initialize()
}

// Holds values needed to set-up connection with NATs
type NatsConfig struct {
	Url        string
	Subject    string
}

func (c *NatsConfig) initialize() {
	log.Println("initializing nats config...")

	cfEnv, err := cfenv.Current()
	if err != nil || cfEnv == nil {
		log.Printf("CF env vars gathering problem: %v (running locally?)", err)
		c.Url = os.Getenv("NATS_URL")
		c.Subject = os.Getenv("NATS_SERVICE_CREATION_SUBJECT")
	} else {
		service, err := cfEnv.Services.WithName("nats-provider")
		if err != nil {
			log.Print("Cannot locate nats service in CF environment")
		}
		c.Url = service.Credentials["url"]
		c.Subject = service.Credentials["service-creation-subject"]
	}
}
