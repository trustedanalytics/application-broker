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

type App struct {
	Guid        string            `json:"guid"`
	Name        string            `json:"name"`
	Environment map[string]string `json:"environment_vars"`
	Space       string            `json:"space"`
	Org         string            `json:"org"`
}

func clientToApp(cfApp cfclient.App) App {
	space := cfApp.Space()
	org := space.Org()
	var app = App{
		Guid:        cfApp.Guid,
		Name:        cfApp.Name,
		Environment: cfApp.Environment,
		Space:       space.Name,
		Org:         org.Name,
	}
	return app
}

func AppHandler(tokens oauth2.Tokens, params martini.Params) []byte {
	config := cfclient.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cfclient.NewClient(&config)
	app := client.AppByGuid(params["app_id"])
	appMarshal, err := json.Marshal(clientToApp(app))
	if err != nil {
		log.Printf("Error marshaling apps %v", err)
	}
	return appMarshal
}

func AppsHandler(tokens oauth2.Tokens) []byte {
	var returnApps []App
	config := cfclient.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cfclient.NewClient(&config)
	apps := client.ListApps()
	for _, app := range apps {
		if app.Environment["APP_LAUNCHER_NAME"] != "" {
			returnApps = append(returnApps, clientToApp(app))
		}
	}

	appsMarshal, err := json.Marshal(returnApps)
	if err != nil {
		log.Printf("Error marshaling apps %v", err)
	}
	return appsMarshal
}
