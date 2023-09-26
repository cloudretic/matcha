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

package tree

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/decentplatforms/matcha/pkg/rctx"
	"github.com/decentplatforms/matcha/pkg/route"
	"github.com/decentplatforms/matcha/pkg/route/require"
)

func TestTree(t *testing.T) {
	rtree := New()
	rtree.Add(route.Declare(http.MethodGet, "/test/route/[p1]{[a-z]*}"))
	rtree.Add(route.Declare(http.MethodGet, "/test/route/[p2]{[a-zA-Z]*}"))
	rtree.Add(route.Declare(http.MethodGet, "/test/route/[ext]+"))
	rtree.Add(route.Declare(http.MethodGet, "/test"))
	rtree.Add(route.Declare(http.MethodGet, "/"))
	req, _ := http.NewRequest(http.MethodGet, "http://test.com/test/route/lowercase", nil)
	req = rctx.PrepareRequestContext(req, rctx.DefaultMaxParams)
	if leaf_id := rtree.Match(req); leaf_id != 1 {
		t.Errorf("wrong match: %d", leaf_id)
	}
	req, _ = http.NewRequest(http.MethodGet, "http://test.com/test/route/Uppercase", nil)
	req = rctx.PrepareRequestContext(req, rctx.DefaultMaxParams)
	if leaf_id := rtree.Match(req); leaf_id != 2 {
		t.Errorf("wrong match: %d", leaf_id)
	}
	req, _ = http.NewRequest(http.MethodGet, "http://test.com/test/route/longer/request", nil)
	req = rctx.PrepareRequestContext(req, rctx.DefaultMaxParams)
	if leaf_id := rtree.Match(req); leaf_id != 3 {
		t.Errorf("wrong match: %d", leaf_id)
	}
	req, _ = http.NewRequest(http.MethodGet, "http://test.com/test", nil)
	req = rctx.PrepareRequestContext(req, rctx.DefaultMaxParams)
	if leaf_id := rtree.Match(req); leaf_id != 4 {
		t.Errorf("wrong match: %d", leaf_id)
	}
	req, _ = http.NewRequest(http.MethodGet, "http://test.com/", nil)
	if leaf_id := rtree.Match(req); leaf_id != 5 {
		t.Errorf("wrong match: %d", leaf_id)
	}
	req, _ = http.NewRequest(http.MethodGet, "http://test.com/notfound", nil)
	req = rctx.PrepareRequestContext(req, rctx.DefaultMaxParams)
	if leaf_id := rtree.Match(req); leaf_id != NO_LEAF_ID {
		t.Error("route shouldn't exist")
	}
	req, _ = http.NewRequest(http.MethodPost, "http://test.com/test/route/lowercase", nil)
	req = rctx.PrepareRequestContext(req, rctx.DefaultMaxParams)
	if leaf_id := rtree.Match(req); leaf_id != NO_LEAF_ID {
		t.Error("route shouldn't exist")
	}
}

func TestDuplicate(t *testing.T) {
	rtree := New()
	a := rtree.Add(route.Declare(http.MethodGet, "/duplicate/route"))
	b := rtree.Add(route.Declare(http.MethodGet, "/duplicate/route"))
	if a != 1 {
		t.Errorf("expected leaf_id 1, got %d", a)
	}
	if b != 2 {
		t.Errorf("expected leaf_id 2, got %d", b)
	}
	c := rtree.Match(httptest.NewRequest(http.MethodGet, "/duplicate/route", nil))
	if c != 1 {
		t.Errorf("expected leaf_id 1, got %d", c)
	}
}

func TestRequire(t *testing.T) {
	rtree := New()
	a := rtree.Add(route.Declare(http.MethodGet, "/", route.Require(require.Hosts("test.com"))))
	b := rtree.Add(route.Declare(http.MethodGet, "/"))
	if a != 1 {
		t.Errorf("expected leaf_id 1, got %d", a)
	}
	if b != 2 {
		t.Errorf("expected leaf_id 2, got %d", b)
	}
	c := rtree.Match(httptest.NewRequest(http.MethodGet, "http://test.com/", nil))
	if c != 1 {
		t.Errorf("expected leaf_id 1, got %d", c)
	}
	c = rtree.Match(httptest.NewRequest(http.MethodGet, "/", nil))
	if c != 2 {
		t.Errorf("expected leaf_id 2, got %d", c)
	}
}
