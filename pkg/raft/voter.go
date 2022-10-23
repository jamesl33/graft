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

// RequestVoteInput encapsulates the request options available for the 'RequestVote' API.
type RequestVoteInput struct {
	// Term is the candidate’s term.
	Term Term `json:"term"`

	// CandidateID is the id of the candidate requesting the vote.
	CandidateID Peer `json:"candidate_id"`

	// LastLogIndex is the index of candidate’s last log entry.
	LastLogIndex int `json:"last_log_index"`

	// LastLogTerm is the term of candidate’s last log entry.
	LastLogTerm Term `json:"last_log_term"`
}

// RequestVoteOutput encapsulates the response options for the 'RequestVote' API.
type RequestVoteOutput struct {
	// Term is the current term, for candidate to update itself.
	Term Term `json:"term"`

	// Granted is set to true to indicate that the candidate received the vote.
	Granted bool `json:"granted"`
}

// Voter is the API for requesting a vote from a node in the cluster.
//
//  1. On conversion to candidate, start election:
//
//     - Increment currentTerm.
//
//     - Vote for self.
//
//     - Reset election timer.
//
//     - Send RequestVote RPCs to all other servers.
//
//  2. If votes received from majority of servers: become leader
//
//  3. If AppendEntries RPC received from new leader: convert to follower
//
//  4. If election timeout elapses: start new election
type Voter interface {
	// RequestVote requests a vote from the node.
	//
	// 1. Reply false if term < currentTerm.
	// 2. If votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant
	//    vote.
	RequestVote(input RequestVoteInput) (output RequestVoteOutput, err error)
}
