package fctest

import (
	"encoding/binary"

	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/eth2/forkchoice"
)

func LighthouseTestDef() *ForkChoiceTestDef {
	spec := configs.Mainnet
	hash := func(i uint64) (out forkchoice.Root) {
		binary.LittleEndian.PutUint64(out[:8], i)
		return
	}
	//epoch2Slot := func(epoch Epoch) Slot {
	//	s, _ := spec.EpochStartSlot(epoch)
	//	return s
	//}
	init := ForkChoiceTestInit{
		Spec:         spec,
		Finalized:    forkchoice.Checkpoint{Root: hash(0), Epoch: 0},
		Justified:    forkchoice.Checkpoint{Root: hash(0), Epoch: 0},
		AnchorRoot:   hash(0),
		AnchorSlot:   0,
		AnchorParent: hash(0),
		Balances:     []forkchoice.Gwei{spec.MAX_EFFECTIVE_BALANCE, spec.MAX_EFFECTIVE_BALANCE},
	}
	var ops []Operation
	add := func(op Operation) {
		ops = append(ops, op)
	}

	// Ensure that the head starts at the finalized block.
	add(&OpHead{
		ExpectedHead: forkchoice.NodeRef{Root: hash(0), Slot: 0},
		Ok:           true,
	})

	// Add a block with a hash of 2, at slot 2
	//
	//          0
	//          |
	//          *
	//         /
	//        2
	add(&OpProcessBlock{
		Parent:         hash(0),
		BlockRoot:      hash(2),
		BlockSlot:      2,
		JustifiedEpoch: 0,
		FinalizedEpoch: 0,
	})

	// Ensure that the head is 2
	//
	//          0
	//          |
	//          *
	//         /
	// head-> 2
	add(&OpHead{
		ExpectedHead: forkchoice.NodeRef{Root: hash(2), Slot: 2},
		Ok:           true,
	})

	// Add a block with a hash of 1 that comes off the genesis block (this is a fork compared
	// to the previous block). At slot 1, it arrived late.
	//
	//          0
	//         / \
	//        *   1
	//        |
	//        2
	add(&OpProcessBlock{
		Parent:         hash(0),
		BlockRoot:      hash(1),
		BlockSlot:      1,
		JustifiedEpoch: 0,
		FinalizedEpoch: 0,
	})

	// Ensure that the head is 2 (tie-break on higher root)
	//
	//          0
	//         / \
	//        *   1
	//        |
	//        2
	add(&OpHead{
		ExpectedHead: forkchoice.NodeRef{Root: hash(2), Slot: 2},
		Ok:           true,
	})

	// Add a vote to block 2
	add(&OpProcessAttestation{
		ValidatorIndex: 0,
		BlockRoot:      hash(2),
		HeadSlot:       2,
		CanAdd:         true,
	})

	// Ensure that the head is now 2
	add(&OpHead{
		ExpectedHead: forkchoice.NodeRef{Root: hash(2), Slot: 2},
		Ok:           true,
	})

	// TODO: many more steps

	return &ForkChoiceTestDef{
		Init:       init,
		Operations: ops,
	}
}
