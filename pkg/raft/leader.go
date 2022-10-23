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

// SetInput encapsulates the request options for the 'Set' API.
type SetInput struct {
	// Key is the key being mutated.
	Key string `json:"key"`

	// Value is the new value that will be associated with the key.
	Value any `json:"value"`
}

// GetInput encapsulates the request options for the 'Get' API.
type GetInput struct {
	// Key is the get being fetched.
	Key string `json:"key"`
}

// GetOutput encapsulates the response options for the 'Get' API.
type GetOutput struct {
	// Key is the key that was fetched.
	Key string `json:"key"`

	// Value is the value associated with the key.
	Value any `json:"value"`
}

// DeleteInput encapsulates the request options for the 'Delete' API.
type DeleteInput struct {
	// Key is the key being deleted.
	Key string `json:"key"`
}

// Leader is the API for servicing client requests to the replicated log.
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
type Leader interface {
	// Set a key/value pair.
	Set(input SetInput) (err error)

	// Get a key value pair.
	Get(input GetInput) (output GetOutput, err error)

	// Delete a key.
	Delete(input DeleteInput) (err error)
}
