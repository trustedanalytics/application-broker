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

package client

import (
	log "github.com/cihub/seelog"
	"github.com/cloudfoundry-community/go-cfenv"
)

type AppDependencyDiscovererUPS struct {
	Url      string `json:"url"`
	AuthUser string `json:"auth_user"`
	AuthPass string `json:"auth_pass"`
}

func NewAppDependencyDiscovererUPS(envs *cfenv.App) *AppDependencyDiscovererUPS {
	appDepDiscUps := new(AppDependencyDiscovererUPS)
	appDepDiscUps.Url = "http://localhost:9998"
	if envs == nil {
		log.Warnf("CF Env vars do not exist. Using %v as app dependency discoverer url",
			appDepDiscUps.Url)
		return appDepDiscUps
	}
	appDepUps, err := envs.Services.WithName("app-dependency-discoverer-ups")
	if err != nil {
		log.Warnf("app-dependency-discoverer-ups not defined. Using %v as app dependency discoverer url",
			appDepDiscUps.Url)
		return appDepDiscUps
	}
	if url, ok := appDepUps.Credentials["url"]; ok {
		appDepDiscUps.Url = url.(string)
	}
	if auth_user, ok := appDepUps.Credentials["auth_user"]; ok {
		appDepDiscUps.AuthUser = auth_user.(string)
	}
	if auth_pass, ok := appDepUps.Credentials["auth_pass"]; ok {
		appDepDiscUps.AuthPass = auth_pass.(string)
	}
	return appDepDiscUps
}
