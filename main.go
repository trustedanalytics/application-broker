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

// Package main application-broker API
//
// The purpose of this application is to easily expose Cloud Foundry applications to marketplace as service offerings
// without implementing separate brokers for each of them.
//
//     Version: 0.7.3
//
// swagger:meta
package main

import (
	log "github.com/cihub/seelog"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/trustedanalytics/application-broker/broker"
	"github.com/trustedanalytics/application-broker/cloud"
	"github.com/trustedanalytics/application-broker/dao"
	"github.com/trustedanalytics/application-broker/logging"
	"github.com/trustedanalytics/application-broker/messagebus"
	"github.com/trustedanalytics/application-broker/service"
)

func main() {

	var mbus messagebus.MessageBus
	var err error

	logging.Initialize()

	cfEnv, err := cfenv.Current()
	if err != nil {
		log.Warnf("CF Env vars gathering failed with error [%v]. Running locally, probably.", err)
	}
	natsConfig := messagebus.Config{}
	natsAvailable := natsConfig.TryInitialize(cfEnv)

	if natsAvailable {
		if mbus, err = messagebus.NewNatsMessageBus(natsConfig); err != nil {
			log.Warnf("Failed to initialize NATS. Events information publishing will be skipped. [%v]", err)
			mbus = &messagebus.DevNullBus{}
		}
	} else {
		log.Warn("NATS not available. Events information publishing will be skipped.")
		mbus = &messagebus.DevNullBus{}
	}

	db := dao.MongoFactory(cfEnv)
	cloud := cloud.NewCloudAPI(cfEnv)
	s := service.New(db, cloud, mbus, service.CreationStatusFactory{})

	b, err := broker.New(s)
	if err != nil {
		log.Criticalf("failed to initialize broker: [%v]", err)
	}

	brokerCfg := broker.Config{}
	brokerCfg.Initialize(cfEnv)
	b.Start(brokerCfg)
}
