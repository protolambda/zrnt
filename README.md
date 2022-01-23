# ZRNT [![Go Report Card](https://goreportcard.com/badge/github.com/protolambda/zrnt?no-cache)](https://goreportcard.com/report/github.com/protolambda/zrnt) [![CircleCI Build Status](https://circleci.com/gh/protolambda/zrnt.svg?style=shield)](https://circleci.com/gh/protolambda/zrnt) [![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

A minimal Go implementation of the Eth 2 spec, by @protolambda.

The goal of this project is to have a Go version of the Python based spec,
 to enable use cases that are out of reach for unoptimized Python.

## Structure of `eth2` package

### `beacon`

The beacon package covers the Beacon Chain spec, but optimized for performance.

It is split into packages per fork (`phase0`, `altair`, `bellatrix`, etc.),
that all follow the same functions/naming patterns.

Globals are split up as following:
- `<something>Type`: a ZTYP SSZ type description, used to create typed views with. Some types can be derived from a `*Spec` only.
- `<something>View`: a ZTYP SSZ view. This wraps a binary-tree backing,
 to provide typed a mutable interface to the persistent cached binary-tree datastructure.

Backing all the forks, the `common` package implements common types and transition functionality:
- Basic types
- Beacon block headers
- Fork-agnostic Beacon block envelopes
- `StateTransition`, `ProcessSlots`, and misc. transition base functions
- `Spec`, the standard eth2 configuration, parametrizes a lot of the beacon functionality.

The genesis is implemented in `phase0`, but forks may also implement additional genesis variants, to start a genesis into the fork.

The `Spec` type is very central, and enables multiple different spec configurations at the same time, as well as full customization of the configuration.
Default configs are available in the `configs` package: `Mainnet` and `configs.Minimal`: preconfigured `*Spec`s to use for common tooling/tests.
Alternatively, load phase0 and phase1 configs in `Spec` from YAML, and then embed them in a `Spec` object to use them.

#### Working around lack of Generics

A common pattern is to have `As<Something` functions that work like `(view View, err error) -> (typedView *SomethingView, err error)`.
Since Go has no generics, ZTYP does not deserialize/get any typed views. If you load a new view, or get an attribute,
it will have to be wrapped with an `As<Something>` function to be fully typed. The 2nd `err` input is for convenience:
chaining getters and `As` functions together is much easier this way. If there is any error, it is proxied.

### `chain`

The chain is split in two interfaces:
- `ColdChain`, implemented by `FinalizedChain`: a linear series of slot and block transitions.
- `HotChain`, implemented by `UnfinalizedChain`: a tree of slots and blocks, backed by the forkchoice graph.

The current `UnfinalizedChain` keeps all states in memory. This seems like a lot of state, but the state data uses data-sharing to avoid any duplication.
The result is that it merely keeps some 1 state and some diffs in memory, and this is performant when switching between many hot-states, as everything is in memory already, no disk-IO!
However, for prolonged non-finalizing chains (e.g. no finalization for more than a day), the memory can become a problem.
The state persistence trade-off here is being weighed against the complexity drawbacks.

The `FullChain` interfaces combines the two into a usable eth2 chain, where blocks and attestations can be added to, and the canonical chain can be determined and navigated.

### `configs`

The common configurations are already included by default, no need to add or run anything if you just need `mainnet` or `minimal` spec.

For custom configurations, simply load the `beacon.Phase0Config` and/or `beacon.Phase1Config` from YAML, and put them in the `beacon.Spec` object.

BLS can be turned off on compile-time by adding the `bls_off` build tag (security warning: for testing use only!).

### `db`

This package offers Blocks and States DB implementations, to simply store and retrieve the common consensus data. 

### `forkchoice`

Forkchoice consists of 3 parts:
- The `Forkchoice` interface, the wrapper around all internals, thread safe.
- The `ForkchoiceGraph` interface and `ProtoArray` implementation: efficiently track and update the DAG with LMD-GHOST forkchoice rules.
- The `VoteStore` interface and `ProtoVoteStore` implementation: track the latest votes and weight of each validator, to compute batched diffs as votes change, to then apply to the `ForkchoiceGraph`.

The forkchoice implements block-slot accuracy voting. The internal representation tracks two different graphs:
- Transition graph: slot processing and block processing are sequential
- Forkchoice graph: slot votes and block votes are contentious (slots count as empty blocks)

The transition graph is used for navigation, and allows for efficient state building (no repeated epoch transitions),
while the forkchoice graph accurately follows voting edge cases such as for gap slot heads.

The forkchoice implementation is undergoing more testing and may not be completely stable.

### `pool`

Implements in-memory collections for the common gossip message topics:
- Attestations: in aggregated and individual form
- Attester Slashings
- Proposer Slashings
- Voluntary Exits

### `util`

Hashing, merkleization, and other utils can be found in `eth2/util`.

SSZ is provided by ZTYP, but has two forms:
- Native Go structs, with `Deserialize`, `Serialize` and `HashTreeRoot` methods, built on the `hashing` and `codec` packages.
- Type descriptions and Views, optimized for caching, representing data as binary trees. Primarily used for the `BeaconStateView`.

### `validation`

This package implements message validation for the Eth2 gossip topics, and requires the chain and blocks DB interfaces to operate.

## Testing

To run all tests and generate test and coverage reports: `make test`

The specs are tested using test-vectors shared between Eth 2.0 clients,
 found here: [`ethereum/eth2.0-spec-tests`](https://github.com/ethereum/eth2.0-spec-tests).
Instructions on the usage of these test-vectors with ZRNT can be found in the [testing readme](./tests/spec/README.md)
 (TLDR: run `make download-tests` first).

### Coverage reports

After running `make test`, run `make open-coverage` to open a Go-test coverage report in your browser.

Note: half of the project consists of `if err != nil` due to tree-structure accesses that do not fit Go error handling well, thus low coverage.

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

