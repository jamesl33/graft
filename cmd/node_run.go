// Copyright 2022 Couchbase Inc.
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

package cmd

import (
	"fmt"

	"github.com/jamesl33/graft/internal/backend/rest"
	"github.com/jamesl33/graft/pkg/raft"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// nodeRunOptions encapsulates the options available for the 'run' command.
var nodeRunOptions = struct {
	// id that will be given to the cluster node.
	id uint

	// port the REST API will run on.
	port uint16

	// peers is an optional list of peers in the cluster.
	peers string
}{}

// nodeRunCommand allows running a cluster node.
var nodeRunCommand = &cobra.Command{
	RunE:  nodeRun,
	Short: "Run a 'graft' cluster node",
	Use:   "run",
}

// init the flags/arguments for the 'run' sub-command.
func init() {
	nodeRunCommand.Flags().UintVar(
		&nodeRunOptions.id,
		"id",
		0,
		"an unsigned integer identifier for the node",
	)

	nodeRunCommand.Flags().Uint16Var(
		&nodeRunOptions.port,
		"port",
		9000,
		"the port used to run the node REST API",
	)

	nodeRunCommand.Flags().StringVar(
		&nodeRunOptions.peers,
		"peers",
		"",
		`and optional list of peers in the format '[{"id":<id>,"address":"<address>"}]'`,
	)

	markFlagRequired(nodeRunCommand, "id")
}

// nodeRun starts up a cluster node.
func nodeRun(_ *cobra.Command, _ []string) error {
	parsedPeers, err := newPeers([]byte(nodeRunOptions.peers))
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal peers")
	}

	convertedPeers := make(map[raft.Peer]raft.Client)

	for _, peer := range parsedPeers {
		convertedPeers[peer.ID] = rest.NewClient(peer.Address)
	}

	node := raft.NewNode(
		raft.Peer(nodeRunOptions.id),
		convertedPeers,
	)

	err = rest.NewServer(node).ListenAndServe(fmt.Sprintf(":%d", nodeRunOptions.port))
	if err != nil {
		return errors.Wrap(err, "failed to list and serve")
	}

	return nil
}
