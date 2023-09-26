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
	"net/http"
)

type Required func(req *http.Request) bool

// ExecuteRequireds executes a list of route validators on a request.
// It only returns true if every validator provided returns true.
func Execute(req *http.Request, vs []Required) bool {
	for _, v := range vs {
		if !v(req) {
			return false
		}
	}
	return true
}
