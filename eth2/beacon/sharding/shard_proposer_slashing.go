package sharding

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var ShardProposerSlashingType = ContainerType("ShardProposerSlashing", []FieldDef{
	{"signed_reference_1", SignedShardBlobReferenceType},
	{"signed_reference_2", SignedShardBlobReferenceType},
})

type ShardProposerSlashing struct {
	SignedReference1 SignedShardBlobReference `json:"signed_reference_1" yaml:"signed_reference_1"`
	SignedReference2 SignedShardBlobReference `json:"signed_reference_2" yaml:"signed_reference_2"`
}

func (v *ShardProposerSlashing) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.SignedReference1, &v.SignedReference2)
}

func (v *ShardProposerSlashing) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.SignedReference1, &v.SignedReference2)
}

func (v *ShardProposerSlashing) ByteLength() uint64 {
	return ShardProposerSlashingType.TypeByteLength()
}

func (*ShardProposerSlashing) FixedLength() uint64 {
	return ShardProposerSlashingType.TypeByteLength()
}

func (v *ShardProposerSlashing) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.SignedReference1, &v.SignedReference2)
}

func BlockShardProposerSlashingsType(spec *common.Spec) ListTypeDef {
	return ListType(ShardProposerSlashingType, spec.MAX_SHARD_PROPOSER_SLASHINGS)
}

type ShardProposerSlashings []ShardProposerSlashing

func (a *ShardProposerSlashings) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, ShardProposerSlashing{})
		return &((*a)[i])
	}, ShardProposerSlashingType.TypeByteLength(), spec.MAX_SHARD_PROPOSER_SLASHINGS)
}

func (a ShardProposerSlashings) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, ShardProposerSlashingType.TypeByteLength(), uint64(len(a)))
}

func (a ShardProposerSlashings) ByteLength(*common.Spec) (out uint64) {
	return ShardProposerSlashingType.TypeByteLength() * uint64(len(a))
}

func (a *ShardProposerSlashings) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li ShardProposerSlashings) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, spec.MAX_SHARD_PROPOSER_SLASHINGS)
}

func ProcessShardProposerSlashings(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state common.BeaconState, ops []ShardProposerSlashing) error {
	for i := range ops {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := ProcessShardProposerSlashing(spec, epc, state, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func ProcessShardProposerSlashing(spec *common.Spec, epc *common.EpochsContext, state common.BeaconState, proposerSlashing *ShardProposerSlashing) error {
	ref1 := &proposerSlashing.SignedReference1.Message
	ref2 := &proposerSlashing.SignedReference2.Message

	// Verify header slots match
	if ref1.Slot != ref2.Slot {
		return fmt.Errorf("invalid shard proposer slashing, slots must be equal, got %d <> %d", ref1.Slot, ref2.Slot)
	}
	// Verify header shards match
	if ref1.Shard != ref2.Shard {
		return fmt.Errorf("invalid shard proposer slashing, shards must be equal, got %d <> %d", ref1.Shard, ref2.Shard)
	}
	// Verify header proposer indices match
	if ref1.ProposerIndex != ref2.ProposerIndex {
		return fmt.Errorf("invalid shard proposer slashing, proposers must be equal, got %d <> %d", ref1.ProposerIndex, ref2.ProposerIndex)
	}
	// Verify the headers are different (i.e. different body)
	if ref1.BodyRoot == ref2.BodyRoot {
		return fmt.Errorf("invalid shard proposer slashing, body roots must be different, got %s <> %s", ref1.BodyRoot, ref2.BodyRoot)
	}
	// Verify the proposer is slashable
	validators, err := state.Validators()
	if err != nil {
		return err
	}
	validator, err := validators.Validator(ref1.ProposerIndex)
	if err != nil {
		return err
	}
	if slashable, err := phase0.IsSlashable(validator, epc.CurrentEpoch.Epoch); err != nil {
		return err
	} else if !slashable {
		return fmt.Errorf("shard proposer slashing requires proposer (%d) to be slashable", ref1.ProposerIndex)
	}
	domain, err := common.GetDomain(state, common.DOMAIN_BEACON_PROPOSER, spec.SlotToEpoch(ref1.Slot))
	if err != nil {
		return err
	}
	pubkey, ok := epc.PubkeyCache.Pubkey(ref1.ProposerIndex)
	if !ok {
		return fmt.Errorf("could not find pubkey of proposer %d", ref1.ProposerIndex)
	}
	// Verify signatures
	if !bls.Verify(
		pubkey,
		common.ComputeSigningRoot(ref1.HashTreeRoot(tree.GetHashFn()), domain),
		proposerSlashing.SignedReference1.Signature) {
		return errors.New("shard proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.Verify(
		pubkey,
		common.ComputeSigningRoot(ref2.HashTreeRoot(tree.GetHashFn()), domain),
		proposerSlashing.SignedReference2.Signature) {
		return errors.New("shard proposer slashing header 2 has invalid BLS signature")
	}
	return nil
}
