package eth2

type Slot uint64
type Epoch uint64
type Shard uint64
type Gwei uint64
type Timestamp uint64
type ValidatorIndex uint64
type DepositIndex uint64
type BLSDomain uint64

// byte arrays
type Root [32]byte
type Bytes32 [32]byte
type BLSPubkey [48]byte
type BLSSignature [96]byte

type ValueFunction func(index ValidatorIndex) Gwei


func (s Slot) ToEpoch() Epoch {
	return Epoch(s / SLOTS_PER_EPOCH)
}

func (e Epoch) GetStartSlot() Slot {
	return Slot(e) * SLOTS_PER_EPOCH
}

