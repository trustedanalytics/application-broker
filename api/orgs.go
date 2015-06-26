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
	"encoding/json"
	"log"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/oauth2"
)

func OrgsHandler(tokens oauth2.Tokens) []byte {
	config := cfclient.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cfclient.NewClient(&config)
	orgs := client.ListOrgs()
	orgsMarshal, err := json.Marshal(orgs)
	if err != nil {
		log.Printf("Error marshaling orgs %v", err)
	}
	return orgsMarshal
}

func OrgSpaceHandler(tokens oauth2.Tokens, params martini.Params) []byte {
	config := cfclient.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cfclient.NewClient(&config)
	spaces := client.OrgSpaces(params["org_id"])
	spacesMarshal, err := json.Marshal(spaces)
	if err != nil {
		log.Printf("Error marshaling spaces %v", err)
	}
	return spacesMarshal
}
