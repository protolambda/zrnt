# Spec-testing

## Installing

To run the spec tests, you need the test-vectors provided here: https://github.com/ethereum/eth2.0-spec-tests
These vectors are hosted in a [Git LFS](https://git-lfs.github.com/) repository. 
Alternatively, you can download a `.tar.gz` from the releases page.

Next, you place the repository in `<zrnt root>/tests/spec/eth2.0-spec-tests`, or, symlink to them (your paths may vary):
```bash
ln -s ../../../eth2.0-specs/eth2.0-spec-tests tests/spec/eth2.0-spec-tests
```

## Running

The tests are compatible with the go test tooling,
 and can be ran per-handler, and are filtered based on build settings.

Order:
1. Test runner collection
    Example: `go test -tags preset_minimal ./test_runners/...`
2. Test runner: e.g. operations
    Example: `go test -tags preset_minimal ./test_runners/operations`
3. Test handler: e.g. transfer
    Run: `go test -tags preset_minimal ./test_runners/operations -run TestTransfer`
4. Test suite(s) loaded by handler, filtered based on active configuration (same build tags apply to tests)
    To filter, use the title of the suite (can be incomplete / match multiple).
    Example: `go test -tags preset_minimal ./test_runners/operations -run 'TestTransfer/transfer operation`
5. Test cases, executed by handler
    To filter, use the title provided by the test. or the number (if the handler does not provide titling).
    Examples: 
    - `go test -tags preset_minimal ./test_runners/operations -run 'TestTransfer/transfer operation/success_withdrawable'`
    - `go test -tags preset_minimal ./test_runners/ssz_static -run 'TestSSZStatic/chaos/case #3'`
6. Test case conditions
    A test fails/pass, withs stacktrace etc. 
    Go test tooling docs here:
    [`cmd`](https://golang.org/cmd/go/#hdr-Test_packages) and [`testing pkg`](https://golang.org/pkg/testing/)

