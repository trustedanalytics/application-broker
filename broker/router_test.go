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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Router", func() {

	var (
		sut *router
	)

	Describe("after instantiation", func() {
		It("should be set to use Basic Authentication", func() {
			recorder := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/v2/catalog", nil)

			sut = newRouter(nil)
			sut.ServeHTTP(recorder, r)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})
