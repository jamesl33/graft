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
	"context"
	"net/url"
)

// Client is a client allowing a node to communicate to other nodes in the cluster.
type Client interface {
	// Address returns the address client is using to communicate to the node.
	Address() *url.URL

	// AppendEntries is invoked by leader to replicate log entries (§5.3); also used as heartbeat.
	AppendEntries(ctx context.Context, input AppendEntriesInput) (output AppendEntriesOutput, err error)

	// RequestVote is invoked by candidates to gather votes.
	RequestVote(ctx context.Context, input RequestVoteInput) (output RequestVoteOutput, err error)
}
