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

package service

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestConfig(t *testing.T) {

	assert.NotEmpty(t, Config, "nil config")
	assert.NotNil(t, Config.CatalogPath, "nil catalog path")
	assert.NotNil(t, Config.Dependencies, "nil deps")
	assert.Equal(t, 2, len(Config.Dependencies), "incorrect number of deps")

	for i, dep := range Config.Dependencies {
		log.Printf("dep[%d]:%s (%s)", i, dep.Name, dep.Plan)
		assert.NotNil(t, dep.Name, "nil name")
		assert.NotNil(t, dep.Plan, "nil plan")
	}

	assert.NotNil(t, Config.CFEnv, "nil CFEnv")

}
