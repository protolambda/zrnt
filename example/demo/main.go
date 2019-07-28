package main

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	"math/rand"
)

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
				Amount:                MAX_EFFECTIVE_BALANCE,
				Signature:             BLSSignature{1, 2, 3}, // BLS not yet implemented
			},
		}
		depLeafHash := ssz.HashTreeRoot(&dep.Data, DepositDataSSZ)
		deposits = append(deposits, dep)
		depRoots = append(depRoots, depLeafHash)
	}
	hashFn := htr.HashFn(hashing.GetHashFn())
	leaf := func(i uint64) []byte {
		return depRoots[i][:]
	}
	for i := uint64(0); i < uint64(len(deposits)); i++ {
		proof := merkle.ConstructProof(hashFn, i+1, 1<<DEPOSIT_CONTRACT_TREE_DEPTH, leaf, i)
		for j := 0; j < DEPOSIT_CONTRACT_TREE_DEPTH; j++ {
			copy(deposits[i].Proof[j][:], proof[j][:])
		}
		binary.LittleEndian.PutUint64(deposits[i].Proof[DEPOSIT_CONTRACT_TREE_DEPTH][:], i+1)
	}

	state, err := GenesisFromEth1(Root{42}, genesisTime, deposits)
	if err != nil {
		panic(err)
	}

	full := &FullFeatures{}
	blockProc := &BlockProcessFeature{}
	for i := 0; i < 300; i++ {
		block := SimulateBlock(state, rng)
		full.Load(state)
		blockProc.Block = block
		blockProc.Meta = full
		if err := full.StateTransition(blockProc, false); err != nil {
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
			RandaoReveal: BLSSignature{4, 2},
			Eth1Data: Eth1Data{
				DepositRoot: Root{0, 1, 3},
				BlockHash:   Root{4, 5, 6},
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
