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

package api

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
)

// CfAPI is the implementation of API interface. It is point of access to CF CloudController API
type CfAPI struct {
	BaseAddress string
	*http.Client
}

// NewCfAPI constructs and initializes access to CF by loading necessary credentials from ENVs
func NewCfAPI() *CfAPI {
	envs := cfenv.CurrentEnv()
	tokenConfig := &clientcredentials.Config{
		ClientID:     envs["CLIENT_ID"],
		ClientSecret: envs["CLIENT_SECRET"],
		Scopes:       []string{},
		TokenURL:     envs["TOKEN_URL"],
	}
	toReturn := new(CfAPI)
	toReturn.BaseAddress = envs["CF_API"]
	toReturn.Client = tokenConfig.Client(oauth2.NoContext)
	return toReturn
}
