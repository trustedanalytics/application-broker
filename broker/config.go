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
	Debug        bool
	CFEnv        *cfenv.App
}

func (c *BrokerConfig) initialize() {
	log.Println("initializing broker config...")
	c.Debug = os.Getenv("CF_DEBUG") == "true"

	cfEnv, err := cfenv.Current()
	if err != nil || cfEnv == nil {
		log.Printf("failed to get CF env vars, probably running locally: %v", err)
		cfEnv = &cfenv.App{}
		cfEnv.Port = service.GetEnvVarAsInt("PORT", 9999)
		cfEnv.Host = "0.0.0.0"
		cfEnv.TempDir = os.TempDir()
	}
	c.CFEnv = cfEnv

	c.validate()
}

func (c *BrokerConfig) validate() {
	missingEnvVars := []string{}
	if len(missingEnvVars) > 0 {
		log.Println("Missing environment variable configuration:")
		for _, envVar := range missingEnvVars {
			log.Printf("* %s", envVar)
		}
		os.Exit(1)
	}
}
