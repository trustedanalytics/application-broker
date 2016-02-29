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

package wrapper

import (
	"github.com/trustedanalytics/go-cf-lib/types"
)

func (w *CfAPIWrapper) RegisterBroker(brokerName string, brokerURL string, username string, password string) error {
	return w.rest.RegisterBroker(brokerName, brokerURL, username, password)
}

func (w *CfAPIWrapper) UpdateBroker(brokerGUID string, brokerURL string, username string, password string) error {
	return w.rest.UpdateBroker(brokerGUID, brokerURL, username, password)
}

func (w *CfAPIWrapper) GetBrokers(brokerName string) (*types.CfServiceBrokerResources, error) {
	brokers, err := w.rest.GetBrokers(brokerName)
	return brokers, err
}
