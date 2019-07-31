package epoch_processing

import (
	"github.com/protolambda/messagediff"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/zssz"
	"testing"
)

type TestCase struct {
	Pre  *phase0.BeaconState
	Post *phase0.BeaconState
}

func (c *TestCase) TestCondition(t *testing.T) {
	// in case hashes are incorrectly correct (e.g. new SSZ behavior), we still have diffs
	if diff, equal := messagediff.PrettyDiff(c.Pre, c.Post, messagediff.SliceWeakEmptyOption{}); !equal {
		t.Fatalf("end result does not match expectation!\n%s", diff)
	}
}

func LoadTest(t *testing.T, readPart test_util.TestPartReader) (out TestCase) {
	r, size := readPart("pre.ssz")
	test_util.Check(t, zssz.Decode(r, size, out.Pre, phase0.BeaconStateSSZ))
	r, size = readPart("post.ssz")
	test_util.Check(t, zssz.Decode(r, size, out.Post, phase0.BeaconStateSSZ))
	return
}

func MakeRunner(runTestCase func(t *testing.T, testCase TestCase)) test_util.CaseRunner {
	return test_util.HandleBLS(func(t *testing.T, readPart test_util.TestPartReader) {
		testCase := LoadTest(t, readPart)
		runTestCase(t, testCase)
	})
}
