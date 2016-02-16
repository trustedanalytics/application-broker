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
package misc

import (
	"bytes"
	"crypto/rand"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/nu7hatch/gouuid"
	"io"
	"os"
	"strconv"
	"strings"
)

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

func ReduceInstanceID(id string) string {
	idSplit := strings.Split(id, "-")
	return strings.Join(idSplit[0:len(idSplit)-1], "-")
}

func NewGUID() string {
	uuid, _ := uuid.NewV4()
	return uuid.String()
}

func FirstNonEmpty(elems chan error, size int) error {
	for i := 0; i < size; i++ {
		if elem := <-elems; elem != nil {
			return elem
		}
	}
	return nil
}

func ReaderToString(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String()
}

func GenerateRandomString(length int) string {
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, length)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

func ReplaceWithRandom(value string) string {
	for i := 8; i <= 32; i += 8 {
		for strings.Contains(value, fmt.Sprintf("$RANDOM%d", i)) {
			value = strings.Replace(value, fmt.Sprintf("$RANDOM%d", i), GenerateRandomString(i), 1)
		}
	}
	return value
}
