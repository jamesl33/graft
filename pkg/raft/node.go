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
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apex/log"
	"github.com/couchbase/tools-common/hofp"
	"github.com/couchbase/tools-common/maths"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

// Node implements the 'Protocol' interface exposing a distributed key/value data store using Raft.
type Node struct {
	// leaderID is the id of the leader node, zero if unknown.
	leaderID Peer

	// id is the node id.
	id Peer

	// state is the current state i.e. follower, candidate or leader.
	state State

	// term is the latest term server has seen (initialized to 0 on first boot, increases monotonically)
	term Term

	// votedFor is the candidateId that received vote in current term (or zero if none)
	votedFor Peer

	// log containing etries with commands for state machine, and term when entry was received by leader (zero indexed)
	log Log

	// commitIndex is the index of highest log entry known to be committed (initialized to 0, increases monotonically)
	commitIndex int

	// peers is a list of peer ids i.e. the nodes in the cluster.
	peers []Peer

	// clients used to connect to each peer.
	clients map[Peer]Client

	// nextIndexes is a map containing the index of the next log entry to send to a peer (initialized to leader last log
	// index + 1)
	//
	// NOTE: Reinitialised upon election as leader.
	nextIndexes map[Peer]int

	// matchIndexes is a map containing the index of highest log entry known to be replicated on a peer (initialized to
	// 0, increases monotonically)
	//
	// NOTE: Reinitialised upon election as leader.
	matchIndexes map[Peer]int

	// mu guards access to internal state.
	mu sync.Mutex

	// updatedCommitIndex signifies that the commit index has been updated.
	updatedCommitIndex *sync.Cond

	// triggerReplication allows triggering of replication after a call to 'Set'.
	//
	// NOTE: Use a 'Cond' not a channel to avoid potential deadlocks due to state changes. Internally this is converted
	// into a signal channel where required.
	triggerReplication *sync.Cond

	// lastElectionEvent is the time at which the election event occured (i.e. to defer an election from taking place)
	lastElectionEvent time.Time
}

// NewNode creates a node with the given id, connected to the provided peers.
func NewNode(id Peer, peers map[Peer]Client) *Node {
	srv := Node{
		id:                 id,
		state:              StateFollower,
		log:                make(Log, 0),
		commitIndex:        -1,
		peers:              maps.Keys(peers),
		clients:            make(map[Peer]Client),
		updatedCommitIndex: sync.NewCond(&sync.Mutex{}),
		triggerReplication: sync.NewCond(&sync.Mutex{}),
	}

	for peer, client := range peers {
		srv.clients[peer] = client
	}

	go srv.start()

	return &srv
}

// start up the node as a follower.
func (n *Node) start() {
	log.WithFields(log.Fields{"id": n.id}).Info("Starting node")

	for {
		switch n.getState() {
		case StateFollower:
			n.runElectionTimer()
		case StateCandidate:
			n.campaign()
		case StateLeader:
			n.lead()
		}
	}
}

// getState returns the current state.
//
// NOTE: This method is thread safe.
func (n *Node) getState() State {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.state
}

// getTerm returns the current term.
//
// NOTE: This method is thread safe.
func (n *Node) getTerm() Term {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.term
}

// lastIndexAndTerm returns the index of the last entry in the log and the term it was created in.
//
// NOTE: This method is thread safe.
func (n *Node) lastIndexAndTerm() (int, Term) {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.lastIndexAndTermLocked()
}

// lastIndexAndTermLocked returns the index of the last entry in the log and the term it was created in.
func (n *Node) lastIndexAndTermLocked() (int, Term) {
	if len(n.log) == 0 {
		return -1, 0
	}

	idx := len(n.log) - 1

	return idx, n.log[idx].Term
}

// transitionLOCKED transitions the node into the given state and sets the term to that provided.
func (n *Node) transitionLocked(state State, term Term) {
	fields := log.Fields{
		"id":    n.id,
		"state": state,
		"term":  term,
	}

	log.WithFields(fields).Info("Transitioning state")

	n.leaderID = 0 // Unknown
	n.state = state
	n.term = term

	switch state {
	case StateFollower:
		n.votedFor = 0
	case StateCandidate:
		n.votedFor = n.id
	case StateLeader:
		n.leaderID = n.id
		n.nextIndexes = make(map[Peer]int)
		n.matchIndexes = make(map[Peer]int)

		var idx int
		if len(n.log) != 0 {
			idx = len(n.log)
		}

		for _, peer := range n.peers {
			n.nextIndexes[peer] = idx
		}
	}

	n.lastElectionEvent = time.Now()
}

// runElectionTimer starts a timer which transitions the node to a candidate if no messages are received from the leader
// node.
func (n *Node) runElectionTimer() {
	const (
		min = 150 * time.Millisecond
		max = 300 * time.Millisecond
	)

	var (
		term    = n.getTerm()
		timeout = time.Duration(rand.Int63n(int64(max)-int64(min)) + int64(min))
	)

	<-time.After(timeout)

	n.mu.Lock()
	defer n.mu.Unlock()

	if !(n.state == StateFollower || n.state == StateCandidate) || n.term != term {
		return
	}

	if time.Since(n.lastElectionEvent) < timeout {
		return
	}

	n.transitionLocked(StateCandidate, n.term+1)
}

// campaign triggers a leadership election where the current node is the candidate.
func (n *Node) campaign() {
	var (
		term         = n.getTerm()
		votes uint64 = 1
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := hofp.NewPool(hofp.Options{
		Context: ctx,
		Size:    len(n.peers) + 1, // +1 for the election timer
	})

	// Start the election timer, if we don't hear back from the other nodes another election should be triggered
	_ = pool.Queue(func(_ context.Context) error { n.runElectionTimer(); return nil })

	// handle the request vote output.
	handle := func(output RequestVoteOutput) {
		n.mu.Lock()
		defer n.mu.Unlock()

		switch {
		case n.state != StateCandidate:
			// No longer a candidate, ignore the response
		case output.Term > term:
			// Other node is further ahead, become a follower
			n.transitionLocked(StateFollower, output.Term)
		case output.Term == term && output.Granted && atomic.AddUint64(&votes, 1)*2 > uint64(len(n.peers)+1):
			// Received a majority of votes, become the leader
			n.transitionLocked(StateLeader, n.term)
		}
	}

	vote := func(ctx context.Context, client Client) {
		lastIndex, lastTerm := n.lastIndexAndTerm()

		input := RequestVoteInput{
			Term:         term,
			CandidateID:  n.id,
			LastLogIndex: lastIndex,
			LastLogTerm:  lastTerm,
		}

		output, err := client.RequestVote(ctx, input)
		switch err {
		case nil:
			handle(output)
		default:
			fields := log.Fields{
				"address": client.Address().String(),
				"error":   err.Error(),
			}

			log.WithFields(fields).Debugf("Failed to request vote")
		}
	}

	queue := func(client Client) error {
		return pool.Queue(func(ctx context.Context) error { vote(ctx, client); return nil })
	}

	for _, client := range n.clients {
		if queue(client) != nil {
			break
		}
	}

	_ = pool.Stop()
}

// lead begins heartbeating/log replicating.
func (n *Node) lead() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Convert the 'triggerReplication' cond into a signal channel to allow use in a select case
	//
	// NOTE: The channel will be closed by 'replicationTrigger'.
	signal := make(chan struct{}, 1)
	go n.replicationTrigger(ctx, signal)

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if n.getState() != StateLeader {
			return
		}

		n.replicate()

		select {
		case <-signal:
		case <-ticker.C:
		}
	}
}

// replicationTrigger sends on the given channel whenever replication is manually triggered after a write.
func (n *Node) replicationTrigger(ctx context.Context, signal chan<- struct{}) {
	defer close(signal)

	n.triggerReplication.L.Lock()
	defer n.triggerReplication.L.Unlock()

	for {
		n.triggerReplication.Wait()

		select {
		case <-ctx.Done():
			return
		case signal <- struct{}{}:
		}
	}
}

// replicate replicates the log to the other cluster nodes (and acts as a heartbeat).
func (n *Node) replicate() {
	var (
		term = n.getTerm()
		stop = errors.New("stop") // Cancel replication early, used only as a sentinel error
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := hofp.NewPool(hofp.Options{
		Context: ctx,
		Size:    len(n.peers),
	})

	replicate := func(ctx context.Context, peer Peer) error {
		if n.replicateToPeer(ctx, term, peer) {
			return stop
		}

		return nil
	}

	queue := func(peer Peer) error {
		return pool.Queue(func(ctx context.Context) error { return replicate(ctx, peer) })
	}

	for _, peer := range n.peers {
		if queue(peer) != nil {
			break
		}
	}

	_ = pool.Stop()
}

// replicateToPeer replicates the log to the given peer.
func (n *Node) replicateToPeer(ctx context.Context, term Term, peer Peer) bool {
	n.mu.Lock()

	var (
		nextIndex = n.nextIndexes[peer]
		prevIndex = nextIndex - 1
		prevTerm  Term
	)

	if prevIndex >= 0 {
		prevTerm = n.log[prevIndex].Term
	}

	// Copy the entries being replicated
	entries := make([]Entry, len(n.log[nextIndex:]))
	copy(entries, n.log[nextIndex:])

	input := AppendEntriesInput{
		Term:             term,
		ID:               n.id,
		PreviousLogIndex: prevIndex,
		PreviousLogTerm:  prevTerm,
		Entries:          entries,
		CommitIndex:      n.commitIndex,
	}

	n.mu.Unlock()

	client, ok := n.clients[peer]
	if !ok {
		log.Fatalf("invalid state, no client for peer %d", peer)
	}

	// Ignore failures to replicate log, we'll try again later
	output, ok := n.performAppendEntries(ctx, client, input)
	if !ok {
		return false
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// Other node is further ahead, become a follower and stop replication
	if output.Term > n.term {
		n.transitionLocked(StateFollower, output.Term)
	}

	// No longer the leader, stop replication (not just required for the above state transition)
	if n.state != StateLeader {
		return true
	}

	// Not the expected term, ignore
	if output.Term != term {
		return false
	}

	// Peer is further behind, decrement the next index; next time replication is triggered more entries will be sent
	//
	//nolint:godox
	// TODO(jamesl33): Implement optimization described in section 5.3 of https://raft.github.io/raft.pdf.
	if !output.Success {
		n.nextIndexes[peer] = nextIndex - 1
		return false
	}

	n.nextIndexes[peer] = nextIndex + len(input.Entries)
	n.matchIndexes[peer] = n.nextIndexes[peer] - 1

	if !n.updateCommitIndex() {
		return false
	}

	log.WithField("index", n.commitIndex).Info("Updated commit index")

	// Wake up any goroutines waiting for replication before responding to a 'Set'
	n.updatedCommitIndex.Broadcast()

	return false
}

// performAppendEntries uses the given client to perform the append entries request, logs in the event of a failure.
func (n *Node) performAppendEntries(
	ctx context.Context,
	client Client,
	input AppendEntriesInput,
) (AppendEntriesOutput, bool) {
	output, err := client.AppendEntries(ctx, input)
	if err == nil {
		return output, true
	}

	fields := log.Fields{
		"address": client.Address().String(),
		"error":   err.Error(),
	}

	log.WithFields(fields).Debugf("Failed to append entries")

	return AppendEntriesOutput{}, false
}

// updateCommitIndex updates the commit index when log entries have been replicated to a quorum of nodes.
func (n *Node) updateCommitIndex() bool {
	saved := n.commitIndex

	for idx := saved + 1; idx < len(n.log); idx++ {
		if n.log[idx].Term != n.term || !n.replicatedToMajorityLocked(idx) {
			continue
		}

		n.commitIndex = idx
	}

	return n.commitIndex != saved
}

// Set creates a new entry in the log and triggers replication.
func (n *Node) Set(input SetInput) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state != StateLeader {
		return n.notLeaderLocked()
	}

	entry := Entry{
		Term:  n.term,
		Key:   input.Key,
		Value: input.Value,
	}

	return n.setLocked(entry)
}

// Get returns an entry from the log if it exists (and is not deleted).
//
// NOTE: System entries will not be returned.
func (n *Node) Get(input GetInput) (GetOutput, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	value, ok := n.log.Get(input.Key, false)
	if !ok || value.Deleted {
		return GetOutput{}, ErrNotFound
	}

	return GetOutput{Key: input.Key, Value: value}, nil
}

// Delete adds an entry to the log indicating the deletion of a key.
//
// NOTE: This is a tombstone; both the entry and the tombstone will remain indefinitely.
func (n *Node) Delete(input DeleteInput) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state != StateLeader {
		return n.notLeaderLocked()
	}

	entry := Entry{
		Term:    n.term,
		Key:     input.Key,
		Deleted: true,
	}

	return n.setLocked(entry)
}

// setLocked adds the given entry to the log and triggers replication.
func (n *Node) setLocked(entry Entry) error {
	log.WithFields(log.Fields{"entry": entry}).Info("Adding entry to log")

	n.log = append(n.log, entry)

	// Trigger immediate replication if it's not already taking place; if it is, replication will be triggered again
	n.triggerReplication.Broadcast()

	for idx := len(n.log) - 1; n.commitIndex < idx; {
		n.updatedCommitIndex.Wait()
	}

	return nil
}

// AppendEntries performs client side log replication and transitions state where required.
//
//nolint:unparam
func (n *Node) AppendEntries(input AppendEntriesInput) (AppendEntriesOutput, error) {
	fields := log.Fields{
		"request": input,
	}

	log.WithFields(fields).Debug("Received append entries request")

	output := n.appendEntries(input)

	fields = log.Fields{
		"response": output,
	}

	log.WithFields(fields).Debug("Dispatching append entries response")

	return output, nil
}

// appendEntries performs client side log replication and transitions state where required.
func (n *Node) appendEntries(input AppendEntriesInput) AppendEntriesOutput {
	n.mu.Lock()
	defer n.mu.Unlock()

	// If AppendEntries RPC received from new leader: convert to follower
	if input.Term > n.term {
		n.transitionLocked(StateFollower, input.Term)
	}

	output := AppendEntriesOutput{
		Term: n.term,
	}

	// Reply false if term < currentTerm
	if input.Term != n.term {
		return output
	}

	if n.state != StateFollower {
		n.transitionLocked(StateFollower, input.Term)
	}

	n.leaderID = input.ID
	n.lastElectionEvent = time.Now()

	var (
		empty = input.PreviousLogIndex == -1
		valid = !empty &&
			input.PreviousLogIndex < len(n.log) && input.PreviousLogTerm == n.log[input.PreviousLogIndex].Term
	)

	// Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm
	if !(empty || valid) {
		return output
	}

	output.Success = true

	// If an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all
	// that follow it and append any new entries not already in the log.
	n.log = n.log.Replicate(input.PreviousLogIndex+1, input.Entries...)

	// if leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if input.CommitIndex > n.commitIndex {
		n.commitIndex = maths.Min(input.CommitIndex, len(n.log)-1)
	}

	return output
}

// RequestVote casts this nodes vote and transitions state where required.
//
//nolint:unparam
func (n *Node) RequestVote(input RequestVoteInput) (RequestVoteOutput, error) {
	fields := log.Fields{
		"request": input,
	}

	log.WithFields(fields).Debug("Received request vote request")

	output := n.requestVote(input)

	fields = log.Fields{
		"response": output,
	}

	log.WithFields(fields).Debug("Dispatching request vote response")

	return output, nil
}

// requestVote casts this nodes vote and transitions state where required.
func (n *Node) requestVote(input RequestVoteInput) RequestVoteOutput {
	n.mu.Lock()
	defer n.mu.Unlock()

	lastIndex, lastTerm := n.lastIndexAndTermLocked()

	if input.Term > n.term {
		n.transitionLocked(StateFollower, input.Term)
	}

	output := RequestVoteOutput{
		Term: n.term,
	}

	var (
		term  = n.term == input.Term
		vote  = n.votedFor == 0 || n.votedFor == input.CandidateID
		valid = input.LastLogTerm > lastTerm || (input.LastLogTerm == lastTerm && input.LastLogIndex >= lastIndex)
	)

	// if votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant vote
	if !(term && vote && valid) {
		return output
	}

	output.Granted = true

	n.votedFor = input.CandidateID

	n.lastElectionEvent = time.Now()

	return output
}

// replicatedToMajorityLocked returns a boolean indicating whether the log up to the given index has been replicated to
// a quorum of nodes.
func (n *Node) replicatedToMajorityLocked(idx int) bool {
	count := 1

	for _, peer := range n.peers {
		if n.matchIndexes[peer] >= idx {
			count++
		}
	}

	return count*2 > len(n.peers)+1
}

// notLeaderLocked returns the most informative error possible indicating this node isn't the leader.
func (n *Node) notLeaderLocked() error {
	client, ok := n.clients[n.leaderID]
	if !ok {
		return ErrLeaderUnknown
	}

	return NotLeaderError{URL: client.Address()}
}
