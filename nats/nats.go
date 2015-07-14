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
package nats

import (
	"log"

	"github.com/nats-io/nats"
)

type MessageBus interface {
	Publish(v interface{})
}

type NatsMessageBus struct {
	NatsConnection      *nats.EncodedConn
}


func NewMessageBus(url string) (*NatsMessageBus, error) {

	log.Printf("creating nats connection: %v", url)
	opts := nats.Options {
		Url: url,
		AllowReconnect: true,
	}

	connection, err := opts.Connect()
	if err != nil {
		log.Printf("Unable to connect with nats: %v", err)
		return nil, err
	}
	log.Println("connection created!")

	encoded, err := nats.NewEncodedConn(connection, nats.JSON_ENCODER)
	if err != nil {
		log.Printf("Unable to create encoded connection with nats: %v", err)
		return nil, err
	}

	return &NatsMessageBus{
		NatsConnection: encoded,
	}, nil
}

func (n *NatsMessageBus) Publish(v interface{}) {
	err := n.NatsConnection.Publish(Config.Subject, v)
	if err != nil {
		log.Panic("Unable to publish message with nats: ", err)
	}
}
