`graft`
======

An implementation of the [Raft](https://raft.github.io/raft.pdf) protocol in Go using a REST API for intra-node
communication.

Building
========

`graft` uses Go modules so may be imported and built using `go`; a makefile is provided for local building and testing.

```sh
$ make       # Build all binaries contained in the ./cmd subdirectory
$ make build # Build all binaries contained in the ./cmd subdirectory
$ make clean # Remove an generated files (build/coverage.out)
```

Development
===========

A convenience makefile is provided for common development tasks.

```sh
$ make coverage # Run unit testing an output a coverage report to 'coverage.out'
$ make generate # Generate mocks
$ make lint     # Run linting
$ make test     # Run unit testing
```

To Do
=====

- [x] Implement leadership election
- [x] Implement log replication
- [x] Add API to set/get/delete key/value data
- [ ] Add a binary/command line user interface
- [ ] Implement persistent storage
- [ ] Implement log compaction/snapshotting
- [ ] Implement node addition/removal
- [ ] Disallow protocol request which originate from outside the cluster
- [ ] Add an RPC/gRPC backend
- [ ] Optimise bulk log replication

FAQ
===

Why REST and not RPC/gRPC?
--------------------------

This was an advanced Go task I set up while at Couchbase for a group new graduate software engineers to get to grips
with Go. I am aware that generally some form of RPC library is used for intra-cluster communication however, given the
graduates would be working on a backend REST API, I valued experience with REST over RPC/gRPC. The method of
intra-cluster communication in `graft` is decoupled meaning additional backends can be added and switched out in the
future.

License
=======

Copyright 2022 James Lee <jamesl33info@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
