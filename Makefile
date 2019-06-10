TEST_OUT_DIR ?= ./test_out

.PHONY: clean create-test-dir test open-coverage

clean:
	rm -rf ./test_out

create-test-dir:
	mkdir -p $(TEST_OUT_DIR)

test: create-test-dir
	gotestsum --junitfile $(TEST_OUT_DIR)/junit.xml -- -tags preset_minimal \
	-coverpkg=github.com/protolambda/zrnt/eth2/... \
	-covermode=count -coverprofile $(TEST_OUT_DIR)/coverage.out \
	./tests/spec/test_runners/... ./eth2/...

open-coverage:
	go tool cover -html=$(TEST_OUT_DIR)/coverage.out

