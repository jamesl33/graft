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
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/mux"
	"github.com/jamesl33/graft/pkg/raft"
	"github.com/pkg/errors"
)

const (
	// API is the root of the REST API.
	API = "/api"

	// APIV1 is the versiond endpoint, all endpoints should be postfixed to this one.
	APIV1 = API + "/v1"

	// APIV1AppendEntries is the endpoint for the 'AppendEntries' operation.
	APIV1AppendEntries = APIV1 + "/append-entries"

	// APIV1RequestVote is the endpoint for the 'RequestVote' operation.
	APIV1RequestVote = APIV1 + "/request-vote"

	// APIV1KV is the endpoint for performing key value operations.
	APIV1KV = APIV1 + "/key-value"
)

// Server implements the Raft client interface allowing intra-cluster communication via a REST API.
type Server struct {
	consensus raft.Protocol
	router    *mux.Router
}

// NewServer returns a server which hosts all the required endpoints to allow intra-cluster communication.
func NewServer(consensus raft.Protocol) *Server {
	svr := Server{
		consensus: consensus,
		router:    mux.NewRouter(),
	}

	svr.router.HandleFunc(APIV1AppendEntries, svr.AppendEntries).Methods(http.MethodPost)
	svr.router.HandleFunc(APIV1RequestVote, svr.RequestVote).Methods(http.MethodPost)
	svr.router.HandleFunc(APIV1KV, svr.Set).Methods(http.MethodPost)
	svr.router.HandleFunc(APIV1KV, svr.Get).Methods(http.MethodGet)
	svr.router.HandleFunc(APIV1KV, svr.Delete).Methods(http.MethodDelete)

	return &svr
}

// ListenAndServe starts the server on the given address/port.
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Handler:      s.router,
		Addr:         addr,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
	}

	return srv.ListenAndServe()
}

// Set a key/value pair.
func (s *Server) Set(writer http.ResponseWriter, req *http.Request) {
	var input raft.SetInput

	if !s.decode(writer, req, &input) {
		return
	}

	err := s.consensus.Set(input)

	switch {
	case errors.Is(err, nil):
		writer.WriteHeader(http.StatusNoContent)
	case s.notLeader(writer, req, err):
	default:
		s.internalError(writer, req, err)
	}
}

// Get returns a key/value pair.
func (s *Server) Get(writer http.ResponseWriter, req *http.Request) {
	var input raft.GetInput

	if !s.decode(writer, req, &input) {
		return
	}

	body, err := process(s.consensus.Get, input)

	switch {
	case errors.Is(err, nil):
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(body)
	case errors.Is(err, raft.ErrNotFound):
		writer.WriteHeader(http.StatusNotFound)
	case s.notLeader(writer, req, err):
	default:
		s.internalError(writer, req, err)
	}
}

// Delete a key/value pair.
func (s *Server) Delete(writer http.ResponseWriter, req *http.Request) {
	var input raft.DeleteInput

	if !s.decode(writer, req, &input) {
		return
	}

	err := s.consensus.Delete(input)

	switch {
	case errors.Is(err, nil):
		writer.WriteHeader(http.StatusNoContent)
	case errors.Is(err, raft.ErrNotFound):
		writer.WriteHeader(http.StatusNotFound)
	case s.notLeader(writer, req, err):
	default:
		s.internalError(writer, req, err)
	}
}

// AppendEntries handles an 'AppendEntries' request.
func (s *Server) AppendEntries(writer http.ResponseWriter, req *http.Request) {
	var input raft.AppendEntriesInput

	if !s.decode(writer, req, &input) {
		return
	}

	body, err := process(s.consensus.AppendEntries, input)
	switch err {
	case nil:
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(body)
	default:
		s.internalError(writer, req, err)
	}
}

// RequestVote handles a 'RequestVote' request.
func (s *Server) RequestVote(writer http.ResponseWriter, req *http.Request) {
	var input raft.RequestVoteInput

	if !s.decode(writer, req, &input) {
		return
	}

	body, err := process(s.consensus.RequestVote, input)
	switch err {
	case nil:
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(body)
	default:
		s.internalError(writer, req, err)
	}
}

// decode the JSON payload from the given request into the provided type.
func (s *Server) decode(writer http.ResponseWriter, req *http.Request, input any) bool {
	err := json.NewDecoder(io.LimitReader(req.Body, 20*1024*1024)).Decode(&input)
	if err == nil {
		return true
	}

	writer.WriteHeader(http.StatusBadRequest)
	_, _ = writer.Write([]byte(http.StatusText(http.StatusBadRequest)))

	return false
}

// notLeader returns a boolean indicating if the given error is due to leadership issues and was handled.
func (s *Server) notLeader(writer http.ResponseWriter, req *http.Request, err error) bool {
	var notLeader raft.NotLeaderError

	switch {
	case errors.Is(err, raft.ErrLeaderUnknown):
		writer.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = writer.Write([]byte(err.Error()))
	case errors.As(err, &notLeader):
		http.Redirect(writer, req, notLeader.URL.JoinPath(APIV1KV).String(), http.StatusTemporaryRedirect)
	default:
		return false
	}

	return true
}

// internalError logs the given error then returns a 500 to the client.
func (s *Server) internalError(writer http.ResponseWriter, req *http.Request, err error) {
	fields := log.Fields{
		"path":  req.URL.Path,
		"error": err.Error(),
	}

	log.WithFields(fields).Error("Internal server error")

	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// process runs the given function with the provided input then marshals the output.
func process[I, O any](fn func(input I) (O, error), input I) ([]byte, error) {
	output, err := fn(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute function")
	}

	body, err := json.Marshal(output)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return body, nil
}
