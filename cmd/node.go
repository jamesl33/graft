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
	"github.com/spf13/cobra"
)

// nodeCommand is a command to allow running/managing cluster nodes.
var nodeCommand = &cobra.Command{
	Short: "Run/manage 'graft' cluster nodes",
	Use:   "node",
}

// init the node command by adding all the supported sub-commands.
func init() {
	nodeCommand.AddCommand(nodeRunCommand)
}
