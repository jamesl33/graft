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

// Protocol is a definition of the Raft consensus protocol for managing a replicated log.
//
// All Nodes:
//
//  1. If commitIndex > lastApplied: increment lastApplied, apply log[lastApplied] to state machine.
//  2. If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower.
//
// Followers:
//
//  1. Respond to RPCs from candidates and leaders
//  2. If election timeout elapses without receiving AppendEntries RPC from current leader or granting vote to
//     candidate: convert to candidate
//
// Candidates:
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
//
// Leaders:
//
//  1. Upon election: send initial empty AppendEntries RPCs (heartbeat) to each server; repeat during idle periods to
//     prevent election timeouts.
//
//  2. If command received from client: append entry to local log, respond after entry applied to state machine.
//
//  3. If last log index ≥ nextIndex for a follower: send AppendEntries RPC with log entries starting at nextIndex:
//
//     - If successful: update nextIndex and matchIndex for follower.
//
//     - If AppendEntries fails because of log inconsistency: decrement nextIndex and retry.
//
//  4. If there exists an N such that N > commitIndex, a majority of matchIndex[i] ≥ N, and log[N].term == currentTerm:
//     set commitIndex = N.
type Protocol interface {
	Follower
	Voter
	Leader
}
