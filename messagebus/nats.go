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

package messagebus

import (
	log "github.com/cihub/seelog"

	"github.com/nats-io/nats"
)

// NatsMessageBus is an implementation of MessageBus interface
type NatsMessageBus struct {
	NatsConnection *nats.EncodedConn
	Subject        string
}

// NewNatsMessageBus is constructor for nats connection wrapper
func NewNatsMessageBus(configuration Config) (MessageBus, error) {

	log.Debugf("creating nats connection: %v", configuration.url)
	connection, err := nats.Connect(configuration.url)
	if err != nil {
		log.Debugf("Unable to connect with nats: [%v]", err)
		return nil, err
	}
	log.Debug("connection created!")

	encoded, err := nats.NewEncodedConn(connection, nats.JSON_ENCODER)
	if err != nil {
		log.Debugf("Unable to create encoded connection with nats: [%v]", err)
		return nil, err
	}

	return &NatsMessageBus{
		NatsConnection: encoded,
		Subject:        configuration.subject,
	}, nil
}

// Publish sends given message to the bus
func (n *NatsMessageBus) Publish(m Message) {
	err := n.NatsConnection.Publish(n.Subject, m)
	if err != nil {
		log.Errorf("Unable to publish message with nats: [%v]", err)
	}
}
