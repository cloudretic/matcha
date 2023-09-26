// Copyright 2023 Matcha Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package require

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetReqHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "www.test.com"
	host, port := getReqHost(req)
	if host != "www.test.com" || port != "80" {
		t.Error(host, port)
	}
	req.Host = "www.test.com:8080"
	host, port = getReqHost(req)
	if host != "www.test.com" || port != "8080" {
		t.Error(host, port)
	}
	req.Host = "www.invalid.com:8080:8081"
	host, port = getReqHost(req)
	if host != "" || port != "" {
		t.Error(host, port)
	}
}

func TestRequireHosts(t *testing.T) {
	rq := Hosts("localhost", "{.+}.decentplatforms.com")
	// Positive cases
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3000", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "http://localhost:4500", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "https://www.decentplatforms.com:443", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "https://api.decentplatforms.com:443", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "https://api.decentplatforms.com", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "http://api.decentplatforms.com", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	// Negative cases
	req = httptest.NewRequest(http.MethodGet, "https://decentplatforms.com", nil)
	if rq(req) {
		t.Error("expected no match")
	}
}

func TestRequireHostPorts(t *testing.T) {
	rq := HostPorts("localhost:3000", "localhost:3001-4000,4500", "https://{.+}.decentplatforms.com")
	// Positive cases
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3000", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	for i := 3001; i <= 4000; i++ {
		url := fmt.Sprintf("http://localhost:%d", i)
		req = httptest.NewRequest(http.MethodGet, url, nil)
		if !rq(req) {
			t.Error("expected match", url)
		}
	}
	req = httptest.NewRequest(http.MethodGet, "http://localhost:4500", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "https://www.decentplatforms.com:443", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "https://api.decentplatforms.com:443", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	req = httptest.NewRequest(http.MethodGet, "https://api.decentplatforms.com", nil)
	if !rq(req) {
		t.Error("expected match")
	}
	// Negative cases
	req = httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	if rq(req) {
		t.Error("expected no match")
	}
	req = httptest.NewRequest(http.MethodGet, "http://api.decentplatforms.com", nil)
	if rq(req) {
		t.Error("expected no match")
	}

	// Failure cases
	// The only valid port here is 8021.
	rq = HostPorts("test.com:8000a,8001a-8010,8011-8020a,8021")
	req = httptest.NewRequest(http.MethodGet, "http://test.com:8000", nil)
	if rq(req) {
		t.Error("expected no match")
	}
	req = httptest.NewRequest(http.MethodGet, "http://test.com:8005", nil)
	if Execute(req, []Required{rq}) {
		t.Error("expected no match")
	}
	req = httptest.NewRequest(http.MethodGet, "http://test.com:8021", nil)
	if !Execute(req, []Required{rq}) {
		t.Error("expected match")
	}
}
