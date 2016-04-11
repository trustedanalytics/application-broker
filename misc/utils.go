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
	"crypto/rand"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"strings"
)

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
