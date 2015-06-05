/**
 * Copyright (c) 2015 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	UI     bool
	Debug  bool
}

func (c *APIConfig) initialize() {
	log.Println("initializing broker config...")

	c.ApiURL = os.Getenv("API_URL")
	c.UI = os.Getenv("UI") == "true"

	c.validate()
}

func (c *APIConfig) validate() {
	missingEnvVars := []string{}

	if c.UI {
		if c.ApiURL == "" {
			missingEnvVars = append(missingEnvVars, "API_URL")
		}
	}
	if len(missingEnvVars) > 0 {
		log.Println("Missing environment variable configuration:")
		for _, envVar := range missingEnvVars {
			log.Printf("* %s", envVar)
		}
		os.Exit(1)
	}
}
