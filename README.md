# ZRNT [![Go Report Card](https://goreportcard.com/badge/github.com/protolambda/zrnt?no-cache)](https://goreportcard.com/report/github.com/protolambda/zrnt) [![CircleCI Build Status](https://circleci.com/gh/protolambda/zrnt.svg?style=shield)](https://circleci.com/gh/protolambda/zrnt) [![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

A minimal Go implementation of the ETH 2.0 spec, by @protolambda.

The goal of this project is to have a Go version of the Python based spec,
 to enable use cases that are out of reach for unoptimized Python.

Think of:
- Realtime use in test-network monitoring, fast enough to verify full network activity.
- A much higher fuzzing speed, to explore for new bugs.
- Not affected nearly as much as pyspec when scaling to mainnet.

## Structure of `eth2` package

### `beacon`

The beacon package covers the phase0 state transition, following the Beacon-chain spec, but optimized for performance.

#### Globals

Globals are split up as following:
- `<something>Type`: a ZTYP SSZ type description, used to create typed views with. Some types can be derived from a `*Spec` only.
- `<something>View`: a ZTYP SSZ view. This wraps a binary-tree backing,
 to provide typed a mutable interface to the persistent cached binary-tree datastructure.
- `(spec *Spec) Process<something>`: The processing functions are all prefixed, and attached to the beacon-state view.
- `(spec *Spec) StateTransition`: The main transition function
- `EpochsContext` and `(spec *Spec) NewEpochsContext`: to efficiently transition, data that is used accross the epoch slot is cached in a special container.
 This includes beacon proposers, committee shuffling, pubkey cache, etc. Functions to efficiently copy and update are included.
- `(spec *Spec) GenesisFromEth1` and `KickStartState` two functions to create a `*BeaconStateView` and accompanying `*EpochsContext` with.

The `Spec` type is very central, and enables multiple different spec configurations at the same time, as well as full customization of the configuration.
Default configs are available in the `configs` package: `Mainnet` and `configs.Minimal`: preconfigured `*Spec`s to use for common tooling/tests.
Alternatively, load phase0 and phase1 configs in `Spec` from YAML, and then embed them in a `Spec` object to use them.

#### Working around lack of Generics

A common pattern is to have `As<Something` functions that work like `(view View, err error) -> (typedView *SomethingView, err error)`.
Since Go has no generics, ZTYP does not deserialize/get any typed views. If you load a new view, or get an attribute,
it will have to be wrapped with an `As<Something>` function to be fully typed. The 2nd `err` input is for convenience:
chaining getters and `As` functions together is much easier this way. If there is any error, it is proxied.

#### Importing

If you work with many of the types and functions, it may be handy to import with a `.` for brevity.

```go
import . "github.com/protolambda/zrnt/eth2/beacon"
```

### `util`

Hashing, merkleization, and other utils can be found in `eth2/util`.

BLS is a package that wraps Herumi BLS. However, it is put behind a build-flag. Use `bls_on` and `bls_off` to use it or not.

SSZ is provided by ZTYP, but has two forms:
- Native Go structs, with `Deserialize`, `Serialize` and `HashTreeRoot` methods, built on the `hashing` and `codec` packages.
- Type descriptions and Views, optimized for caching, representing data as binary trees. Primarily used for the `BeaconStateView`.

### Building

ZRNT is a library, to use it in action, try [ZCLI](github.com/protolambda/zcli) or [Rumor](github.com/protolambda/rumor).

#### Customizing configuration

The common configurations are already included by default, no need to add or run anything if you just need `mainnet` or `minimal` spec.

For custom configurations, add one or more configuration files to `./presets/configs/`.
Then, re-generate the dynamic parts by running `go generate ./...`.

#### Choosing a config

By default, the `defaults.go` preset is used, which is selected if no other presets are,
i.e. `!preset_mainnet,!preset_minimal` if `mainnet` and `minimal` are the others.

To make a specific configuration effective, add a build-constraint (also known as "build tag") when building ZRNT or using it as a dependency:

```shell script
# Examples, Go commands are all consistent, something like:
go test -tags preset_minimal ./...
go build -tags preset_minimal ./...
```

BLS can be turned off by adding the `bls_off` build tag (security warning: for testing use only!).

### Testing

To run all tests and generate test and coverage reports: `make test`

The specs are tested using test-vectors shared between Eth 2.0 clients,
 found here: [`ethereum/eth2.0-spec-tests`](https://github.com/ethereum/eth2.0-spec-tests).
Instructions on the usage of these test-vectors with ZRNT can be found in the [testing readme](./tests/spec/README.md)
 (TLDR: run `make download-tests` first).

#### Coverage reports

After running `make test`, run `make open-coverage` to open a Go-test coverage report in your browser.

Note: half of the project consists of `if err != nil` due to tree-structure accesses that do not fit Go error handling well, thus low coverage.
[![codecov](https://codecov.io/gh/protolambda/zrnt/branch/master/graph/badge.svg?no-cache)](https://codecov.io/gh/protolambda/zrnt) 

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

