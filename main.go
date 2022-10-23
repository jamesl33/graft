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

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jamesl33/graft/cmd"
	"github.com/jamesl33/graft/internal/logging"

	"github.com/apex/log"
	"github.com/pkg/errors"
)

// main will setup logging, then execute graft.
func main() {
	log.SetHandler(logging.NewHandler())

	level, err := log.ParseLevel(os.Getenv("GRAFT_LOG_LEVEL"))
	if err != nil {
		level = log.InfoLevel
	}

	log.SetLevel(level)

	err = cmd.Execute()
	if err == nil {
		return
	}

	// The sub-command failed for some reason, ensure that we exit with a non-zero exit code
	defer os.Exit(1)

	stacktrace := os.Getenv("GRAFT_DISPLAY_STACKTRACE")
	if display, _ := strconv.ParseBool(stacktrace); display {
		fmt.Printf("Error: %+v\n", err)
	} else {
		fmt.Printf("Error: %s\n", errors.Cause(err))
	}
}
