# ZRNT [![Go Report Card](https://goreportcard.com/badge/github.com/protolambda/zrnt?no-cache)](https://goreportcard.com/report/github.com/protolambda/zrnt) [![CircleCI Build Status](https://circleci.com/gh/protolambda/zrnt.svg?style=shield)](https://circleci.com/gh/protolambda/zrnt) [![codecov](https://codecov.io/gh/protolambda/zrnt/branch/master/graph/badge.svg?no-cache)](https://codecov.io/gh/protolambda/zrnt) [![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

A minimal Go implementation of the ETH 2.0 spec, by @protolambda.

The goal of this project is to have a Go version of the Python based spec,
 to enable use cases that are out of reach for unoptimized Python.

Think of: 
- Realtime use in test-network monitoring, fast enough to verify full network activity.
- A much higher fuzzing speed, to explore for new bugs.
- Not affected nearly as much when scaling to mainnet.

## Structure of `eth2`

### `core`

The core package consists of basic data-types, and all constants.
These can be imported with a `.` for brevity, as the name says: these are a core part of the framework ZRNT provides.

```go
import . "github.com/protolambda/zrnt/eth2/core"
```

### `meta`

This package defines the interfaces for `eth2` functionality, some of which know multiple implementations.
These implementations can freely use pre-computed data, or light-client logic, to implement the specific functions.


### `beacon`

The beacon-chain package, the primary focus for Phase-0 of the ETH 2.0 roadmap.

[The beacon-chain specs can be found here.](https://github.com/ethereum/eth2.0-specs/blob/dev/specs/core/0_beacon-chain.md)

Features are split in different packages, and are generally composed as:
- A `State` struct: to embed in a bigger struct, like `phase0.BeaconState`, to implement the given state with.
- A `Feature` struct: combines `meta` information to a feature-specific state, to transition this state, while fully isolated from the other features.
  - This is optional, some features do not require any `meta` information, and consist of just a `State`.

Then, there are several `Status` types, data that is pre-computed by other features, to provide `meta` information with.

### `phase0`

The beacon-chain consists of 3 transition types:

- **Block transition** (occasional, can fork) `eth2/beacon/phase0/block.go > BlockProcessFeature.Process() error`
- **Epoch transition** (every 64 slots, big change) `eth2/beacon/phase0/epoch.go > EpochProcessFeature.Process()`
- **Slot transition** (every slot, small change) `eth2/beacon/phase0/slot.go > SlotProcessFeature.Process()`

The transition between those is standardized in `eth2/beacon/transition/transition.go > StateTransition(block BlockInput, verifyStateRoot bool) error`

The different parts of the transition can be implemented by composing the different `State` and `Feature` structs to provide the necessary `meta` for the transition.
In between forks, this `meta` can be a special implementation to deal with differences. Or a custom implementation to introduce new features and sub-transitions with.

#### Genesis

A genesis state is created with `eth2/beacon/phase0/genesis.go > GenesisFromEth1(eth1BlockHash Root, time Timestamp, deps []Deposit) (*FullFeaturedState, error)`.
This creates a fully featured phase0 state.

From here, a set of slot transitions, block transitions, and epoch transitions, are applied to run the chain.
See `eth2/beacon/transition`.

### `shard`

Planned package, new type of chain introduced in Phase-1.

### `util`

Hashing, merkleization, and other utils can be found in `eth2/util`.

BLS is stubbed, integration of a BLS library is a work in progress.

SSZ is provided by ZRNT-SSZ (ZSSZ), optimized for speed: [`github.com/protolambda/zssz`](https://github.com/protolambda/zssz)

## Getting started

This repo is still a work-in-progress.


### Building

Re-generate dynamic parts by running `go generate ./...` (make sure the presets submodule is pulled).

Add a build-constraint (also known as "build tag") when building ZRNT, or using it as a dependency:

```
go build -tags preset_mainnet

or

go build -tags preset_minimal

```

BLS can be turned off by adding the `bls_off` build tag (security warning: for testing use only!).

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
	return myHashFn                 //  and reset on next call of returned hash-function.
}                                   // A new hash function returned by GetHashFn is never called concurrently.
```

### Conformance testing

Passing all spec-tests is the primary goal to build and direct the Eth 2.0 fuzzing effort.

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

