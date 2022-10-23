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
	"io"
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

// Set dispatches a 'Set' request to the remote node.
func (c *Client) Set(ctx context.Context, input raft.SetInput) error {
	return do(ctx, c.client, http.MethodPost, c.url.JoinPath(APIV1KV), input, nil)
}

// Get dispatches a 'Get' request to the remote node.
func (c *Client) Get(ctx context.Context, input raft.GetInput) (raft.GetOutput, error) {
	var output raft.GetOutput

	return output, do(ctx, c.client, http.MethodGet, c.url.JoinPath(APIV1KV), input, &output)
}

// Delete dispatches a 'Delete' request to the remote node.
func (c *Client) Delete(ctx context.Context, input raft.DeleteInput) error {
	return do(ctx, c.client, http.MethodDelete, c.url.JoinPath(APIV1KV), input, nil)
}

// RequestVote dispatches a request vote request to the remote node.
func (c *Client) RequestVote(ctx context.Context, input raft.RequestVoteInput) (raft.RequestVoteOutput, error) {
	var output raft.RequestVoteOutput

	return output, do(ctx, c.client, http.MethodPost, c.url.JoinPath(APIV1RequestVote), input, &output)
}

// AppendEntries dispatches an append entries request to the remote node.
func (c *Client) AppendEntries(ctx context.Context, input raft.AppendEntriesInput) (raft.AppendEntriesOutput, error) {
	var output raft.AppendEntriesOutput

	return output, do(ctx, c.client, http.MethodPost, c.url.JoinPath(APIV1AppendEntries), input, &output)
}

// do executes a request to the given url using the provided client/input.
func do(
	ctx context.Context,
	client *http.Client,
	method string,
	url *url.URL,
	input interface{},
	output interface{},
) error {
	body, err := json.Marshal(input)
	if err != nil {
		return errors.Wrap(err, "failed to marshal input")
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	if len(body) > 0 && output != nil {
		err = json.Unmarshal(body, &output)
	}

	if err != nil {
		return errors.Wrap(err, "failed to unmarshal output")
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return UnexpectedStatusCodeError{Status: resp.StatusCode, Method: method, Address: url.String(), Body: body}
}
