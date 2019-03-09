# ZRNT

A minimal Go implementation of the ETH 2.0 spec, by @protolambda.

The goal of this project is to have a Go version of the Python based spec:

- executable version
- comparable with Python version (through diffing debug JSON dumps of state)
- generate test vectors for conformance testing
- fast enough for fuzzing
- avoids quadratic complexity implementations of spec
- minimal and easy to understand

## Structure

The beacon-chain consists of 3 transition types:

- **Slot transition** (often, but very minimal change) `eth2/beacon/transition/slot_transition.go`
- **Block transition** (occasional, can fork) `eth2/beacon/transition/block_transition.go`
    `eth2/beacon/block_processing`:
    - Header
    - Randao
    - Eth1
    - ProposerSlashings
    - AttesterSlashings
    - Attestations
    - Deposits
    - VoluntaryExits
    - Transfers
- **Epoch transition** (every 64 slots, big change) `eth2/beacon/transition/epoch_transition.go`
    `eth2/beacon/epoch_processing`:
    - Eth1
    - Crosslinks
    - Crosslinks
    - RewardsAndPenalties (applies deltas) `eth2/beacon/delta_computation`
        - Crosslink rewards/penalties
        - Justificant rewards/penalties
            - normal rewards/penalties
            - delayed finalization measures
    - ValidatorRegistry (shuffling, update)
    - Slashings
    - ExitQueue
    - Finish

Note that sub transitions have a common signature:

Block: `func(state *beacon.BeaconState, block *beacon.BeaconBlock) error` (mutates state, error because they can be rejected)

Epoch: `func(state *beacon.BeaconState)` (mutates state)


A genesis state is created with `eth2/beacon/genesis`.

From here, a set of slot transitions, a block transition, and possibly an epoch transition, are applied for every block `eth2/beacon/transition`.

Data types are defined in `eth2/beacon/data_types`.
Container types are defined in `eth2/beacon/containers`.

SSZ, BLS, hashing, merkleization, and other utils can be found in `eth2/util`

## Getting started

This repo is still a work-in-progress. 
To test creation of some Beacon chain genesis data, you can run `debug/create_debug_json_files.go`

More states and transition testing is coming SOON :tm:


## Future plans

Use the repo for conformance test generation, fuzzing,
 and implement a minimal (but fast) ETH 2.0 client around it.

## Contributing

Contributions are welcome.
If they are not small changes, please make a GH issue and/or contact me first.

## Funding

This project is based on [my work for a bountied challenge by Justin Drake](https://github.com/protolambda/beacon-challenge) to write the ETH 2.0 spec in 1024 lines. A ridiculous challenge, but it was fun, and proved to be useful: 
 every line of code in the spec got extra attention, and it was a fun way to get started on an executable spec.
Aside from this bounty, the work is unfunded (as of now), and developed in my spare time.

## Contact

[@protolambda on Twitter](https://twitter.com/protolambda)

Website: [https://protolambda.com]

## License

MIT, see license file.

