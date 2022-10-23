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

// State is the state of the node i.e. follower, candidate or leader.
type State string

const (
	// StateFollower indicates the node is currently a follower and only responds to requests.
	StateFollower = "follower"

	// StateCandidate indicates the node is currently a candidate and is requesting votes from the other nodes.
	StateCandidate = "candidate"

	// StateLeader indicates the node is the leader and is servicing user requests along with replicating log entires.
	StateLeader = "leader"
)
