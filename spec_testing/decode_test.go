package spec_testing

import (
	"fmt"
	. "github.com/protolambda/zrnt/eth2/util/data_types"
	"strconv"
	"testing"
)

type TestThing struct {
	Foo Root
	Bar Bytes
	Quix uint64
}

func TestDec(t *testing.T) {

	input := map[string]interface{}{
		"Foo": "0x1234567890123456789012345678901234567890123456789012345678901234",
		"Bar": "0xabcd",
		"Quix": strconv.FormatUint(^uint64(0), 10),
	}

	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}