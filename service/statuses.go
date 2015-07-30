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
package service
import "github.com/trustedanalytics/app-launching-service-broker/messagebus"

type ServiceCreationStatus struct {
	ServiceName string
	ServiceType string
	OrgGuid     string
	Message     string
	Timestamp   int64
}

type ServiceCreationStatusFactory struct {
}

func (f ServiceCreationStatusFactory) NewServiceStatus(name string , stype string, org string, msg string) (messagebus.Message) {
	return ServiceCreationStatus{
		ServiceName: name,
		ServiceType: stype,
		OrgGuid: org,
		Message: msg,
		Timestamp: messagebus.GetMillisecondsTimestamp(),
	}
}
