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

package main

// TODO(jamesl33): CLI
// TODO(jamesl33): Add support for clean teardown using context

import (
	"fmt"
	"net/url"
	"os"

	"github.com/apex/log"
	"github.com/jamesl33/graft/internal/backend/rest"
	"github.com/jamesl33/graft/internal/logging"
	"github.com/jamesl33/graft/pkg/raft"
)

func one() {
	peers := map[raft.Peer]raft.Client{
		2: rest.NewClient(parse("http://127.0.0.1:8082")),
		3: rest.NewClient(parse("http://127.0.0.1:8083")),
	}

	_ = rest.NewServer(raft.NewNode(1, peers)).ListenAndServe(":8081")
}

func two() {
	peers := map[raft.Peer]raft.Client{
		1: rest.NewClient(parse("http://127.0.0.1:8081")),
		3: rest.NewClient(parse("http://127.0.0.1:8083")),
	}

	_ = rest.NewServer(raft.NewNode(2, peers)).ListenAndServe(":8082")
}

func three() {
	peers := map[raft.Peer]raft.Client{
		1: rest.NewClient(parse("http://127.0.0.1:8081")),
		2: rest.NewClient(parse("http://127.0.0.1:8082")),
	}

	_ = rest.NewServer(raft.NewNode(3, peers)).ListenAndServe(":8083")
}

func main() {
	log.SetHandler(logging.NewHandler())

	level, err := log.ParseLevel(os.Getenv("GRAFT_LOG_LEVEL"))
	if err != nil {
		level = log.InfoLevel
	}

	log.SetLevel(level)

	one()
}

func parse(host string) *url.URL {
	parsed, err := url.Parse(host)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %q: %v", host, err))
	}

	return parsed
}
