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

// Entry in the replicated log.
type Entry struct {
	// Term is the term when the log entry was applied.
	Term Term `json:"term"`

	// Key is the key being mutated.
	Key string `json:"key"`

	// Value is the value being mutated.
	Value any `json:"value"`

	// System indicates that this is a system event and shouldn't be returned to users.
	//
	//nolint:godox
	// TODO(jamesl33): Use to allow log compaction/node addition/removal.
	System bool `json:"system"`

	// Deleted indicates that this is a deletion.
	Deleted bool `json:"deleted"`
}
