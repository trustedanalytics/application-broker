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

package env

import (
	"bytes"
	"encoding/json"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/go-cf-lib/types"
	"os"
	"strconv"
)

func GetVcapApplication() types.CfVcapApplication {
	vcapSerialized := GetEnvVarAsString("VCAP_APPLICATION", "{}")
	vcap := types.CfVcapApplication{}
	json.NewDecoder(bytes.NewReader([]byte(vcapSerialized))).Decode(&vcap)
	return vcap
}

// ParseInt tries to parse a string into an int; else returns a default value
func ParseInt(s string, defaultInt int) int {
	if len(s) < 1 {
		return defaultInt
	}
	//strconv.Btoi64
	v, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		log.Errorf("unable to parse int from %s: %v", s, err)
		return defaultInt
	}
	return int(v)
}

// GetEnvVarAsString gets an environment variable, or returns a default value if missing/empty
func GetEnvVarAsString(k, defaultEnvValue string) string {
	if len(k) < 1 {
		return defaultEnvValue
	}
	s := os.Getenv(k)
	if len(s) < 1 {
		return defaultEnvValue
	}
	return s
}

// GetEnvVarAsInt gets an env variable and parses to an int; or returns
// a default int if variable missing or not an int
func GetEnvVarAsInt(k string, defaultInt int) int {
	s := GetEnvVarAsString(k, "")
	if len(s) < 1 {
		return defaultInt
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Errorf("unable to parse int from %s: %v", k, err)
		return defaultInt
	}
	return int(v)
}

// GetEnvVarAsBool gets an env variable and parses to a bool; or returns
// a default bool if variable missing or not a bool
func GetEnvVarAsBool(k string, defaultBool bool) bool {
	s := GetEnvVarAsString(k, "")
	if len(s) < 1 {
		return defaultBool
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		log.Errorf("unable to parse bool from %s: %v", k, err)
		return defaultBool
	}
	return v
}
