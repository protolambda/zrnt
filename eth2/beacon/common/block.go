package common

type BeaconBlock interface {
	BlockSlot() Slot
	BlockProposerIndex() ValidatorIndex
	BlockParentRoot() Root
	BlockStateRoot() Root
	Header(spec *Spec) *BeaconBlockHeader
	SpecObj
}

type SignedBeaconBlock interface {
	BlockMessage() BeaconBlock
	BlockSignature() BLSSignature
	SignedHeader(spec *Spec) *SignedBeaconBlockHeader
	SpecObj
}
