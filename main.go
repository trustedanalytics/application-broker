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
package main

import (
	"log"
	"os"

	"github.com/intel-data/app-launching-service-broker/messagebus"
	"github.com/intel-data/app-launching-service-broker/broker"
	"github.com/intel-data/app-launching-service-broker/service"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func main() {

	log.SetFlags(0)

	var n messagebus.MessageBus
	var err error

	natsConfig := messagebus.NatsConfig{}
	natsAvailable := natsConfig.Initialize()
	if natsAvailable {
		n, err = messagebus.NewNatsMessageBus(natsConfig)
	}
	if err != nil || !natsAvailable {
		log.Printf("Failed to initialize nats. Events information publishing will be skipped.")
		n = &messagebus.StubbedNats{}
	}

	s, err := service.New(n)
	if err != nil {
		log.Panicf("failed to initialize service: %v", err)
	}

	b, err := broker.New(s)
	if err != nil {
		log.Panicf("failed to initialize broker: %v", err)
	}

	log.SetOutput(os.Stdout)

	b.Start()
}
