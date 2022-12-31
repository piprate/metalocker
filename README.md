# MetaLocker

[![ci](https://github.com/piprate/metalocker/actions/workflows/ci.yml/badge.svg)](https://github.com/piprate/metalocker/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/piprate/metalocker/branch/main/graph/badge.svg?token=A5Q9M74W5H)](https://codecov.io/github/piprate/metalocker)

MetaLocker is a data distribution framework that enables secure and privacy preserving data exchange between participants of a knowledge network.

It allows users to exchange data with other users in a secure and transparent way while retaining control of their data. MetaLocker protocol is agnostic of communication and storage media, be it a file system, a network storage service or a distributed ledger.

MetaLocker is an open-source version of ChainLocker, a product developed by [Piprate](https://piprate.com), complete with enterprise features like Hyperledger Fabric connector, AWS integration and more.

## Contents

* `cmd` - contains two executables
  * `lockerd` - a reference implementation of MetaLocker architecture
  * `metalo` - a CLI utility to interact with MetaLocker network
* `contexts/files` - JSON-LD contexts that are in use by MetaLocker
* `examples` - a catalog of examples how to interact with MetaLocker
* `index` is a future home for officially supported MetaLocker indexes
  * `bolt` - a BoltDB based MetaLocker index
* `ledger` is a future home for official ledger backends
  * `local` - a BoltDB based MetaLocker ledger
* `model` - core data types and algorithms
* `node` - a basic implementation of MetaLocker node, including a set of API handlers and related infrastructure. `lockerd` relies on this node implementation. You may use it to build your own MetaLocker nodes.
* `remote` - a client library for connecting to MetaLocker nodes.
* `sdk` contains various tools to help building MetaLocker based applications
  * `apibase` - API infrastructure, including authentication handlers
  * `cmdbase` - CLI infrastructure
  * `httpsecure` - a wrapper over HTTP client to support self-signed certs via `https+insecure` schema
  * `testbase` - a test environment that allows spinning off fully-featured, in-memory MetaLocker services for integration testing
* `storage` - a home for official identity backend implementations. These backends store MetaLocker account information
  * `memory` - a transient, in-memory, reference implementation
* `vaults` - a home for official MetaLocker vaults
  * `fs` - a file system based vault for MetaLocker binary artifacts
  * `memory` - a transient, in-memory vault
* `wallet` - a MetaLocker data wallet
