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
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
)

const (
	apiVersion = "v2"
	instanceID = "iid"
	bindingID  = "bid"
)

var (
	catalogURLPattern      = fmt.Sprintf("/%v/catalog", apiVersion)
	provisioningURLPattern = fmt.Sprintf("/%v/service_instances/:instance_id", apiVersion)
	bindingURLPattern      = fmt.Sprintf("/%v/service_instances/:instance_id/service_bindings/:binding_id", apiVersion)
)

type router struct {
	m *martini.ClassicMartini
}

func newRouter(h *handler) *router {

	m := martini.Classic()

	m.Use(sessions.Sessions("app_launcher", sessions.NewCookieStore([]byte("appsecretlauncher"))))
	m.Get(catalogURLPattern, reponseHandler(h.catalog))
	m.Put(provisioningURLPattern, reponseHandler(h.provision))
	m.Delete(provisioningURLPattern, reponseHandler(h.deprovision))
	m.Put(bindingURLPattern, reponseHandler(h.bind))
	m.Delete(bindingURLPattern, reponseHandler(h.unbind))
	return &router{m}
}

// ServeHTTP logs all requests and dispatches to the appropriate handler
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// logging

	if Config.Debug {
		if dump, err := httputil.DumpRequest(req, true); err != nil {
			log.Printf("Cannot log incoming request: %v", err)
		} else {
			log.Print(string(dump))
		}
	}

	if Config.Debug {
		creds, err := parseCredentials(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		//TODO: should not print here, but until implemented
		log.Printf("Router: Authentication: [%v]", creds)
	}

	r.m.ServeHTTP(w, req)
}

type responseEntity struct {
	status int
	value  interface{}
}

type reponseHandler func(*http.Request, martini.Params) (int, string)

// ServeHTTP marshalls response as JSON, return the proper HTTP status code
func (fn reponseHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, params martini.Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fn(req, params)
}

type credentials struct {
	usr string
	pas string
}

func parseCredentials(req *http.Request) (*credentials, error) {
	auths, _ := req.Header["Authorization"]
	if len(auths) != 1 {
		return nil, errors.New("Unauthorized access")
	}
	tokens := strings.Split(auths[0], " ")
	if len(tokens) != 2 || tokens[0] != "bearer" {
		return nil, errors.New("Unsupported authentication method")
	}
	raw, err := base64.StdEncoding.DecodeString(tokens[1])
	if err != nil {
		return nil, errors.New("Unable to decode 'Authorization' header")
	}
	creds := strings.Split(string(raw), ":")
	if len(creds) != 2 {
		return nil, errors.New("Missing credentials")
	}
	return &credentials{creds[0], creds[1]}, nil
}
