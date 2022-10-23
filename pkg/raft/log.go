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
	"github.com/apex/log"
)

// Log is a replicated log where entries contain a command (key/value pair) which is applied to a state machine after
// having been replicated to a quorum of nodes.
type Log []Entry

// Get the latest value for the given key.
func (l Log) Get(key string, system bool) (Entry, bool) {
	for idx := len(l) - 1; idx >= 0; idx-- {
		entry := l[idx]

		if key != entry.Key || (entry.System && !system) {
			continue
		}

		return l[idx], true
	}

	return Entry{}, false
}

// Replicate appends the given entries to the log, truncating any entries which have a mismatched term.
func (l Log) Replicate(prev int, entries ...Entry) Log {
	var idx int

	for {
		if prev+idx >= len(l) || idx >= len(entries) || l[prev+idx].Term != entries[idx].Term {
			break
		}

		idx++
	}

	if len(entries[idx:]) == 0 {
		return l
	}

	fields := log.Fields{
		"entries": entries[idx:],
	}

	log.WithFields(fields).Infof("Replicating entries")

	return append(l[:prev+idx], entries[idx:]...)
}
