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

// AppendEntriesInput encapsulates the request options available for the 'AppendEntries' API.
type AppendEntriesInput struct {
	// Term is the leaders term.
	Term Term `json:"term"`

	// ID is the leaders id so follower can redirect clients.
	ID Peer `json:"id"`

	// PreviousLogIndex is the index of log entry immediately preceding new ones.
	PreviousLogIndex int `json:"previous_log_index"`

	// PreviousLogTerm is the term of prevLogIndex entry.
	PreviousLogTerm Term `json:"previous_log_term"`

	// Entries is a log entries to store (empty for heartbeat; may send more than one for efficiency).
	Entries []Entry `json:"entries"`

	// CommitIndex is the leader’s commitIndex.
	CommitIndex int `json:"commit_index"`
}

// AppendEntriesOutput encapsulates the response options available for the 'AppendEntries' API.
type AppendEntriesOutput struct {
	// Term is the current term, for leader to update itself.
	Term Term `json:"term"`

	// Success is a bollean set to true if follower contained an entry matching prevLogIndex and prevLogTerm.
	Success bool `json:"success"`
}

// Follower is the API for heartbeating/log replication to a node in the cluster.
//
//  1. Respond to RPCs from candidates and leaders
//  2. If election timeout elapses without receiving AppendEntries RPC from current leader or granting vote to
//     candidate: convert to candidate
type Follower interface {
	// AppendEntries heartbeats/replicates the log to the node.
	//
	// 1. Reply false if term < currentTerm.
	// 2. Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm.
	// 3. If an existing entry conflicts with a new one (same index but different terms), delete the existing entry and
	//    all that follow it.
	// 4. Append any new entries not already in the log
	// 5. If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	AppendEntries(input AppendEntriesInput) (output AppendEntriesOutput, err error)
}
