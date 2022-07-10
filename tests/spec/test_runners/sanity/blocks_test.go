package sanity

import (
	"testing"

	"github.com/protolambda/zrnt/tests/spec/test_util"
)

func TestBlocks(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "sanity", "blocks",
		func() test_util.TransitionTest { return new(test_util.BlocksTestCase) })
}
