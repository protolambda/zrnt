TEST_OUT_DIR ?= ./test_out

.PHONY: clean create-test-dir test open-coverage download-tests clear-tests

clean:
	rm -rf ./test_out

create-test-dir:
	mkdir -p $(TEST_OUT_DIR)

SPEC_VERSION ?= v1.1.0-alpha.2

clear-tests:
	rm -rf tests/spec/eth2.0-spec-tests

download-tests:
	mkdir -p tests/spec/eth2.0-spec-tests
	wget https://github.com/ethereum/eth2.0-spec-tests/releases/download/$(SPEC_VERSION)/general.tar.gz -O - | tar -xz -C tests/spec/eth2.0-spec-tests
	wget https://github.com/ethereum/eth2.0-spec-tests/releases/download/$(SPEC_VERSION)/minimal.tar.gz -O - | tar -xz -C tests/spec/eth2.0-spec-tests
	wget https://github.com/ethereum/eth2.0-spec-tests/releases/download/$(SPEC_VERSION)/mainnet.tar.gz -O - | tar -xz -C tests/spec/eth2.0-spec-tests

test: create-test-dir
	gotestsum --junitfile $(TEST_OUT_DIR)/junit.xml -- -tags preset_minimal \
	-coverpkg=github.com/zilm13/zrnt/eth2/... \
	-covermode=count -coverprofile $(TEST_OUT_DIR)/coverage.out \
	./tests/spec/test_runners/... ./eth2/...

open-coverage:
	go tool cover -html=$(TEST_OUT_DIR)/coverage.out

