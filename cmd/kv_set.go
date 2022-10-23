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
	"net/url"

	"github.com/jamesl33/graft/internal/backend/rest"
	"github.com/jamesl33/graft/pkg/raft"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// kvSetCommand allows mutating a key/value pair.
var kvSetCommand = &cobra.Command{
	RunE:  kvSet,
	Short: "Run a 'Set' operation against a 'graft' cluster",
	Use:   "set",
	Args:  cobra.ExactArgs(3),
}

// kvSet uses the REST API to mutate a key/value pair.
func kvSet(_ *cobra.Command, args []string) error {
	parsed, err := url.Parse(args[0])
	if err != nil {
		return errors.Wrap(err, "failed to parse node address")
	}

	err = rest.NewClient(parsed).Set(context.Background(), raft.SetInput{Key: args[1], Value: args[2]})
	if err != nil {
		return errors.Wrap(err, "failed to set key/value pair")
	}

	return nil
}
