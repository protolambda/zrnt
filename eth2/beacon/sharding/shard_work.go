package sharding

import (
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const UNCONFIRMED_SHARD_DATA = 0
const CONFIRMED_SHARD_DATA = 1
const PENDING_SHARD_DATA = 2

func ShardWorkStatusType(spec *common.Spec) *UnionTypeDef {
	return UnionType([]TypeDef{
		nil,
		DataCommitmentType,
		PendingShardHeadersType(spec),
	})
}

type ShardWorkStatus struct {
	Selector uint8 `json:"selector" yaml:"selector"`
	// Either nil, *DataCommitment or *PendingShardHeaders
	Value interface{} `json:"value" yaml:"value"`
}

func (h *ShardWorkStatus) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Union(func(selector uint8) (codec.Deserializable, error) {
		h.Selector = selector
		switch selector {
		case UNCONFIRMED_SHARD_DATA:
			return nil, nil
		case CONFIRMED_SHARD_DATA:
			dat := new(DataCommitment)
			h.Value = dat
			return dat, nil
		case PENDING_SHARD_DATA:
			dat := new(PendingShardHeaders)
			h.Value = dat
			return spec.Wrap(dat), nil
		default:
			return nil, errors.New("bad selector value")
		}
	})
}

func (h *ShardWorkStatus) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	switch h.Selector {
	case UNCONFIRMED_SHARD_DATA:
		return w.Union(UNCONFIRMED_SHARD_DATA, nil)
	case CONFIRMED_SHARD_DATA:
		commitment, ok := h.Value.(*DataCommitment)
		if !ok {
			return fmt.Errorf("invalid value type for CONFIRMED_SHARD_DATA selector: %T", h.Value)
		}
		return w.Union(CONFIRMED_SHARD_DATA, commitment)
	case PENDING_SHARD_DATA:
		headers, ok := h.Value.(*PendingShardHeaders)
		if !ok {
			return fmt.Errorf("invalid value type for PENDING_SHARD_DATA selector: %T", h.Value)
		}
		return w.Union(PENDING_SHARD_DATA, spec.Wrap(headers))
	default:
		return errors.New("bad selector value")
	}
}

func (h *ShardWorkStatus) ByteLength(spec *common.Spec) uint64 {
	if h.Value == nil {
		return 1
	}
	commitment, ok := h.Value.(*DataCommitment)
	if ok {
		return commitment.ByteLength()
	}
	headers, ok := h.Value.(*PendingShardHeaders)
	if !ok {
		return headers.ByteLength(spec)
	}
	return 0
}

func (h *ShardWorkStatus) FixedLength(spec *common.Spec) uint64 {
	return 0
}

func (h *ShardWorkStatus) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	if h.Value == nil {
		return hFn.Union(h.Selector, nil)
	}
	commitment, ok := h.Value.(*DataCommitment)
	if ok {
		return hFn.Union(h.Selector, commitment)
	}
	headers, ok := h.Value.(*PendingShardHeaders)
	if !ok {
		return hFn.Union(h.Selector, spec.Wrap(headers))
	}
	return common.Root{}
}

type ShardWork struct {
	Status ShardWorkStatus `json:"status" yaml:"status"`
}

func (h *ShardWork) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&h.Status))
}

func (h *ShardWork) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&h.Status))
}

func (h *ShardWork) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&h.Status))
}

func (h *ShardWork) FixedLength(spec *common.Spec) uint64 {
	return 0
}

func (h *ShardWork) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&h.Status))
}

func ShardWorkType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("ShardWork", []FieldDef{
		{"status", ShardWorkStatusType(spec)},
	})
}
