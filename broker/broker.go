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
	log "github.com/cihub/seelog"
	"os"
	"os/signal"

	"fmt"
	"github.com/trustedanalytics/application-broker/types"
	"net/http"
)

// Broker represents a running CF Service Broker API
type Broker struct {
	router *router
}

// New creates a loaded instance of the broker
func New(p types.ServiceProviderExtension) (*Broker, error) {
	return &Broker{
		router: newRouter(newHandler(p)),
	}, nil
}

// Start the broker
func (b *Broker) Start(config Config) {

	addr := fmt.Sprintf("%s:%d", config.CFEnv.Host, config.CFEnv.Port)
	log.Infof("starting: %s", addr)

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
		log.Errorf("broker error: %v", err)
	case sig := <-sigCh:
		var _ = sig
		log.Info("broker done")
	}

}
