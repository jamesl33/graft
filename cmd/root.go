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

// rootCommand is the top-level graft command.
var rootCommand = &cobra.Command{
	Short:         "Run/manage 'graft' clusters/nodes",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// init the top-level command by adding all the supported sub-commands.
func init() {
	rootCommand.AddCommand(nodeCommand, kvCommand)
}

// Execute the top-level command.
func Execute() error {
	return rootCommand.Execute()
}
