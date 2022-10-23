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

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/jamesl33/graft/pkg/raft"
	"github.com/pkg/errors"
)

// Client implements the Raft client interface allowing intra-cluster communication via a REST API.
type Client struct {
	url    *url.URL
	client *http.Client
}

// NewClient returns a new client which will communicate with the node a the given address.
func NewClient(url *url.URL) *Client {
	transport := http.Transport{
		DialContext:         (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:   true,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConns:        100,
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Timeout:   50 * time.Millisecond,
		Transport: &transport,
	}

	cli := Client{
		url:    url,
		client: client,
	}

	return &cli
}

// Address returns the address of the remote node.
func (c *Client) Address() *url.URL {
	return c.url
}

// RequestVote dispatches a request vote request to the remote node.
func (c *Client) RequestVote(ctx context.Context, input raft.RequestVoteInput) (raft.RequestVoteOutput, error) {
	return do[raft.RequestVoteInput, raft.RequestVoteOutput](
		ctx,
		c.client,
		http.MethodPost,
		c.url.JoinPath(APIV1RequestVote).String(),
		input,
	)
}

// AppendEntries dispatches an append entries request to the remote node.
func (c *Client) AppendEntries(ctx context.Context, input raft.AppendEntriesInput) (raft.AppendEntriesOutput, error) {
	return do[raft.AppendEntriesInput, raft.AppendEntriesOutput](
		ctx,
		c.client,
		http.MethodPost,
		c.url.JoinPath(APIV1AppendEntries).String(),
		input,
	)
}

// do executes a request to the given url using the provided client/input.
func do[I, O any](ctx context.Context, client *http.Client, method, url string, input I) (O, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return *new(O), errors.Wrap(err, "failed to marhsal input")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return *new(O), errors.Wrap(err, "failed to create request")
	}

	resp, err := client.Do(req)
	if err != nil {
		return *new(O), errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	var output O

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return *new(O), errors.Wrap(err, "failed to decode response")
	}

	return output, nil
}
