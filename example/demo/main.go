package main

import (
	"encoding/binary"
	"github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/beacon/block"
	. "github.com/protolambda/zrnt/eth2/beacon/block/operations"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	"github.com/protolambda/zrnt/eth2/beacon/genesis"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"math/rand"
)

type DepositRoots []Root

func (_ *DepositRoots) Limit() uint32 {
	return 1 << 10 // TODO
}

var DepositRootsSSZ = zssz.GetSSZ((*DepositData)(nil))

func main() {

	// RNG used to create simulated blocks
	rng := rand.New(rand.NewSource(0xDEADBEEF))

	genesisTime := Timestamp(1222333444)
	genesisValidatorCount := 100

	privKeys := make([][32]byte, 0, genesisValidatorCount)
	deposits := make([]Deposit, 0, genesisValidatorCount)
	depRoots := make(DepositRoots, 0, genesisValidatorCount)
	for i := 0; i < genesisValidatorCount; i++ {
		// create a random 32 byte private key. We're not using real crypto yet.
		privKey := [32]byte{}
		rng.Read(privKey[:])
		privKeys = append(privKeys, privKey)
		// simply derive pubkey and withdraw creds, not real thing yet
		pubKey := BLSPubkey{}
		h := hashing.Hash(privKey[:])
		copy(pubKey[:], h[:])
		withdrawCreds := hashing.Hash(append(h[:], 1))
		dep := Deposit{
			Proof: [DEPOSIT_CONTRACT_TREE_DEPTH + 1]Root{},
			Data: DepositData{
				Pubkey:                pubKey,
				WithdrawalCredentials: withdrawCreds,
				Amount:                Gwei(100),
				Signature:             BLSSignature{1, 2, 3}, // BLS not yet implemented
			},
		}
		depLeafHash := ssz.HashTreeRoot(&dep.Data, DepositDataSSZ)
		deposits = append(deposits, dep)
		depRoots = append(depRoots, depLeafHash)
	}
	for i := 0; i < len(deposits); i++ {
		proof := merkle.ConstructProof(depRoots, uint64(i), uint8(DEPOSIT_CONTRACT_TREE_DEPTH))
		copy(deposits[i].Proof[:DEPOSIT_CONTRACT_TREE_DEPTH], proof)
		binary.LittleEndian.PutUint64(deposits[i].Proof[DEPOSIT_CONTRACT_TREE_DEPTH][:], uint64(len(deposits)))
	}

	eth1Data := Eth1Data{
		DepositRoot:  ssz.HashTreeRoot(depRoots, DepositRootsSSZ),
		DepositCount: DepositIndex(len(deposits)),
		BlockHash:    Root{42}, // deposits are simulated, not from a real Eth1 origin.
	}
	state := genesis.Genesis(deposits, genesisTime, eth1Data)

	for i := 0; i < 300; i++ {
		block := SimulateBlock(state, rng)
		err := beacon.StateTransition(state, block, false)
		if err != nil {
			panic(err)
		}
	}

}

func SimulateBlock(state *BeaconState, rng *rand.Rand) *BeaconBlock {
	// copy header
	prevHeader := state.LatestBlockHeader
	// stub state root
	prevHeader.StateRoot = ssz.HashTreeRoot(state, BeaconStateSSZ)
	// get root of previous block
	parentRoot := ssz.HashTreeRoot(prevHeader, BeaconBlockHeaderSSZ)


	block := &BeaconBlock{
		Slot:       state.Slot + 1 + Slot(rng.Intn(5)),
		ParentRoot: parentRoot,
		StateRoot:  Root{},
		Body: BeaconBlockBody{
			RandaoBlockData: RandaoBlockData{RandaoReveal: BLSSignature{4, 2}},
			Eth1BlockData: Eth1BlockData{
				Eth1Data: Eth1Data{
					DepositRoot: Root{0, 1, 3},
					BlockHash:   Root{4, 5, 6},
				},
			},
			Graffiti: Root{123},
			// no operations
		},
		Signature: BLSSignature{1, 2, 3}, // TODO implement BLS
	}
	// TODO: set randao reveal
	// TODO: change eth1 data
	// TODO: sign proposal

	return block
}
