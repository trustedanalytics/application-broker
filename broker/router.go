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
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/sessions"
	"github.com/trustedanalytics/application-broker/env"
	"net/http"
	"net/http/httputil"
)

const (
	apiVersion = "v2"
)

var (
	catalogURLPattern          = fmt.Sprintf("/%v/catalog", apiVersion)
	catalogServiceIdURLPattern = fmt.Sprintf("/%v/catalog/:service_id", apiVersion)
	provisioningURLPattern     = fmt.Sprintf("/%v/service_instances/:instance_id", apiVersion)
	bindingURLPattern          = fmt.Sprintf("/%v/service_instances/:instance_id/service_bindings/:binding_id", apiVersion)
)

type router struct {
	m *martini.ClassicMartini
}

func newRouter(h *handler) *router {

	m := martini.Classic()

	m.Use(auth.Basic(env.GetEnvVarAsString("AUTH_USER", ""), env.GetEnvVarAsString("AUTH_PASS", "")))

	m.Use(sessions.Sessions("app_launcher", sessions.NewCookieStore([]byte("appsecretlauncher"))))
	m.Post(catalogURLPattern, responseHandler(h.append))
	m.Delete(catalogServiceIdURLPattern, responseHandler(h.remove))
	m.Put(catalogServiceIdURLPattern, responseHandler(h.update))
	m.Get(catalogURLPattern, responseHandler(h.catalog))
	m.Put(provisioningURLPattern, responseHandler(h.provision))
	m.Delete(provisioningURLPattern, responseHandler(h.deprovision))
	m.Put(bindingURLPattern, responseHandler(h.bind))
	m.Delete(bindingURLPattern, responseHandler(h.unbind))
	return &router{m}
}

// ServeHTTP logs all requests and dispatches to the appropriate handler
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if dump, err := httputil.DumpRequest(req, true); err != nil {
		log.Tracef("Cannot log incoming request: %v", err)
	} else {
		log.Tracef(string(dump))
	}
	r.m.ServeHTTP(w, req)
}

type responseEntity struct {
	status int
	value  interface{}
}

type responseHandler func(*http.Request, martini.Params) (int, string)

// ServeHTTP marshalls response as JSON, return the proper HTTP status code
func (fn responseHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, params martini.Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fn(req, params)
}
