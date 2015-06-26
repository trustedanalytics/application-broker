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
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/cloudfoundry-community/types-cf"
)

// Broker represents a running CF Service Broker API
type Broker struct {
	router *router
}

// New creates a loaded instance o the broker
func New(p cf.ServiceProvider) (*Broker, error) {
	return &Broker{
		router: newRouter(newHandler(p)),
	}, nil
}

// Start the broker
func (b *Broker) Start() {

	addr := fmt.Sprintf("%s:%d", Config.CFEnv.Host, Config.CFEnv.Port)
	log.Printf("starting: %s", addr)

	sigCh := make(chan os.Signal, 1)

	// make sure we can shutdown gracefully
	signal.Notify(sigCh, os.Interrupt)

	errCh := make(chan error, 1)

	go func() {
		errCh <- http.ListenAndServe(addr, b.router)
	}()

	// non blocking as some of these cf ops are kind of lengthy
	select {
	case err := <-errCh:
		log.Printf("broker error: %v", err)
	case sig := <-sigCh:
		var _ = sig
		log.Print("broker done")
	}

}
