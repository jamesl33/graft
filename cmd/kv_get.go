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
	"context"
	"fmt"
	"net/url"

	"github.com/jamesl33/graft/internal/backend/rest"
	"github.com/jamesl33/graft/pkg/raft"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// kvGetCommand allows fetching a value using its key.
var kvGetCommand = &cobra.Command{
	RunE:  kvGet,
	Short: "Run a 'Get' operation against a 'graft' cluster",
	Use:   "get",
	Args:  cobra.ExactArgs(2),
}

// kvGet uses the REST API to get the value associated with a key.
func kvGet(_ *cobra.Command, args []string) error {
	parsed, err := url.Parse(args[0])
	if err != nil {
		return errors.Wrap(err, "failed to parse node address")
	}

	output, err := rest.NewClient(parsed).Get(context.Background(), raft.GetInput{Key: args[1]})
	if err != nil {
		return errors.Wrap(err, "failed to get key")
	}

	fmt.Println(output.Value)

	return nil
}
