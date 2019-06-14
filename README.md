# ZRNT [![Go Report Card](https://goreportcard.com/badge/github.com/protolambda/zrnt?no-cache)](https://goreportcard.com/report/github.com/protolambda/zrnt) [![CircleCI Build Status](https://circleci.com/gh/protolambda/zrnt.svg?style=shield)](https://circleci.com/gh/protolambda/zrnt) [![codecov](https://codecov.io/gh/protolambda/zrnt/branch/master/graph/badge.svg?no-cache)](https://codecov.io/gh/protolambda/zrnt) [![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

A minimal Go implementation of the ETH 2.0 spec, by @protolambda.

The goal of this project is to have a Go version of the Python based spec:

- Executable version
- Comparable with Python version (through diffing state dumps and running spec-tests)
- fast enough for fuzzing
- avoids inefficiencies of spec
- minimal and easy to understand

## Structure

### Core

The core package consists of basic data-types, and all constants.
These are easily imported with a `.` to write clean code:

```go
import . "github.com/protolambda/zrnt/eth2/core"
```


### Beacon

The beacon-chain package, the primary focus for Phase-0 of the ETH 2.0 roadmap.

[The beacon-chain specs can be found here.](https://github.com/ethereum/eth2.0-specs/blob/dev/specs/core/0_beacon-chain.md)

To use the ZRNT beacon package (with `.` import for minimal code):
```go
import . "github.com/protolambda/zrnt/eth2/beacon"
```

The beacon-chain consists of 2 transition types:

- **Block transition** (occasional, can fork) `eth2/beacon/transition/block_transition.go`
  - Sub transitions: `func(state *beacon.BeaconState, block *beacon.BeaconBlock) error` (mutates state, error because they can be rejected)

- **Epoch transition** (every 64 slots, big change) `eth2/beacon/transition/epoch_transition.go`
   - Sub transitions: `func(state *beacon.BeaconState)` (mutates state)

And, every slot there is a smaller state-change that updates the caches in the state: `eth2/beacon/transition/state_caching.go`


#### Genesis

A genesis state is created with `eth2/beacon/genesis`.

From here, a set of slot transitions, a block transition, and possibly an epoch transition, are applied for every block `eth2/beacon/transition`.

### Shard

Planned package, new type of chain introduced in Phase-1.

### Util

Hashing, merkleization, and other utils can be found in `eth2/util`.

BLS is stubbed, integration of a BLS library is a work in progress.

SSZ is provided by ZRNT-SSZ (ZSSZ), optimized for speed: [`github.com/protolambda/zssz`](https://github.com/protolambda/zssz)

## Getting started

This repo is still a work-in-progress.

More states and transition testing is coming SOON :tm:

### Building

Re-generate dynamic parts by running `go generate ./...` (make sure the presets submodule is pulled).

Add a build-constraint (also known as "build tag") when building ZRNT, or using it as a dependency:

```
go build -tags preset_mainnet

or

go build -tags preset_minimal

etc.
```

### Testing

To run all tests and generate test and coverage reports: `make test`

The specs are tested using test-vectors shared between Eth 2.0 clients,
 found here: [`ethereum/eth2.0-spec-tests`](https://github.com/ethereum/eth2.0-spec-tests).
Instructions on the usage of these test-vectors with ZRNT can be found in the [testing readme](./tests/spec/README.md).

#### Coverage reports

After running `make test`, run `make open-coverage` to open a Go-test coverage report in your browser.

## Future plans

### Fuzzing

There is a fuzzing effort ongoing, based on ZRNT and ZSSZ as base packages for state transitions and encoding.

The fuzzing repo can be found here: https://github.com/guidovranken/eth2.0-fuzzing/.
If you are interested in helping, please open an issue, or contact us through Gitter/Twitter.

#### Replacing the hash-function

For fast fuzzing, the hash-function can be swapped for any other hash function that outputs 32 bytes:

```go
import (
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

// myHashFn: func(input []byte) [32]byte
// e.g. sha256.Sum256

ssz.InitZeroHashes(myHashFn)
hashing.Hash = myHashFn
hashing.GetHashFn = func() HashFn { // Better: re-use allocated working variables,
	return GetHashFn                //  and reset on next call of returned hash-function.
}                                   // A new hash function returned by GetHashFn is never called concurrently.
```

### Conformance testing

Passing all spec-tests is primary goal to build and direct the Eth 2.0 fuzzing effort.

### Minimal client

Besides testing, a Go codebase with full conformance to the eth2.0 spec is also a great play-ground
 to implement a minimal client around, and work on phase-1. If anyone is interested in helping out, contact on Gitter/Twitter.


## Contributing

Contributions are welcome.
If they are not small changes, please make a GH issue and/or contact me first.

## Funding

This project is based on [my work for a bountied challenge by Justin Drake](https://github.com/protolambda/beacon-challenge)
 to write the ETH 2.0 spec in 1024 lines. A ridiculous challenge, but it was fun, and proved to be useful: 
 every line of code in the spec got extra attention, and it was a fun way to get started on an executable spec.
A month later I (@protolambda) started working for the EF,
 and maintain ZRNT to keep up with changes, test the spec, and provide a performant core component for other ETH 2.0 Go projects.

## Contact

Core dev: [@protolambda on Twitter](https://twitter.com/protolambda)

## License

MIT, see license file.

