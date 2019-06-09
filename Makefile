
.PHONY: clean spectest test merge_coverage open_coverage

clean:
	rm -f \
	coverage.txt \
	spectest_coverage.out \
	test_coverage.out

spectest:
	go test -tags preset_minimal \
	-coverpkg=github.com/protolambda/zrnt/eth2/... \
	-covermode=count -coverprofile spectest_coverage.out \
	./tests/spec/test_runners/...

test:
	go test -tags preset_minimal \
	-covermode=count -coverprofile test_coverage.out \
	./eth2/...

merge_coverage:
	cat spectest_coverage.out > coverage.txt
	# ignore the `cover mode` line when merging
	cat test_coverage.out | tail -n +2 >> coverage.txt

open_coverage:
	go tool cover -html=coverage.txt





