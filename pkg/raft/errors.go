// Copyright 2022 James Lee <jamesl33info@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

// ErrLeaderUnknown is returned if the node is not the leader and doesn't know where the leader is.
var ErrLeaderUnknown = errors.New("leader node not found")

// ErrNotFound is returned if something was not found (usually a key).
var ErrNotFound = errors.New("not found")

// NotLeaderError is returned if the node is not the leader but does know where the leader is.
type NotLeaderError struct {
	URL *url.URL
}

// Error implements the 'error' interface returning a useful error indicating whether the leader node is.
func (n NotLeaderError) Error() string {
	return fmt.Sprintf("not the leader, the leader can be found at %q", n.URL)
}
