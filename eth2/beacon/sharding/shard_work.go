package sharding

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
	"gopkg.in/yaml.v3"
)

const SHARD_WORK_UNCONFIRMED = 0
const SHARD_WORK_CONFIRMED = 1
const SHARD_WORK_PENDING = 2

func ShardWorkStatusType(spec *common.Spec) *UnionTypeDef {
	return UnionType([]TypeDef{
		nil,
		AttestedDataCommitmentType,
		PendingShardHeadersType(spec),
	})
}

func AsShardWorkStatus(v View, err error) (*ShardWorkStatusView, error) {
	c, err := AsUnion(v, err)
	return &ShardWorkStatusView{c}, err
}

type ShardWorkStatus struct {
	Selector uint8 `json:"selector" yaml:"selector"`
	// Either nil, *AttestedDataCommitment or *PendingShardHeaders
	Value interface{} `json:"value" yaml:"value"`
}

type selectorJson struct {
	Selector uint8 `json:"selector" yaml:"selector"`
}

type valueJson struct {
	Value interface{} `json:"value" yaml:"value"`
}

func (h *ShardWorkStatus) UnmarshalJSON(b []byte) error {
	var sel selectorJson
	if err := json.Unmarshal(b, &sel); err != nil {
		return err
	}
	var val valueJson
	switch sel.Selector {
	case SHARD_WORK_UNCONFIRMED:
		// nil
	case SHARD_WORK_CONFIRMED:
		val.Value = new(AttestedDataCommitment)
	case SHARD_WORK_PENDING:
		val.Value = new(PendingShardHeaders)
	}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	h.Selector = sel.Selector
	h.Value = val.Value
	return nil
}

func (h *ShardWorkStatus) UnmarshalYAML(value *yaml.Node) error {
	var sel selectorJson
	if err := value.Decode(&sel); err != nil {
		return err
	}
	var val valueJson
	switch sel.Selector {
	case SHARD_WORK_UNCONFIRMED:
		// nil
	case SHARD_WORK_CONFIRMED:
		val.Value = new(AttestedDataCommitment)
	case SHARD_WORK_PENDING:
		val.Value = new(PendingShardHeaders)
	}
	if err := value.Decode(&val); err != nil {
		return err
	}
	h.Selector = sel.Selector
	h.Value = val.Value
	return nil
}

func (h *ShardWorkStatus) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Union(func(selector uint8) (codec.Deserializable, error) {
		h.Selector = selector
		switch selector {
		case SHARD_WORK_UNCONFIRMED:
			return nil, nil
		case SHARD_WORK_CONFIRMED:
			dat := new(AttestedDataCommitment)
			h.Value = dat
			return dat, nil
		case SHARD_WORK_PENDING:
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
	case SHARD_WORK_UNCONFIRMED:
		return w.Union(SHARD_WORK_UNCONFIRMED, nil)
	case SHARD_WORK_CONFIRMED:
		attested, ok := h.Value.(*AttestedDataCommitment)
		if !ok {
			return fmt.Errorf("invalid value type for SHARD_WORK_CONFIRMED selector: %T", h.Value)
		}
		return w.Union(SHARD_WORK_CONFIRMED, attested)
	case SHARD_WORK_PENDING:
		headers, ok := h.Value.(*PendingShardHeaders)
		if !ok {
			return fmt.Errorf("invalid value type for SHARD_WORK_PENDING selector: %T", h.Value)
		}
		return w.Union(SHARD_WORK_PENDING, spec.Wrap(headers))
	default:
		return errors.New("bad selector value")
	}
}

func (h *ShardWorkStatus) ByteLength(spec *common.Spec) uint64 {
	switch h.Selector {
	case SHARD_WORK_UNCONFIRMED:
		return 1
	case SHARD_WORK_CONFIRMED:
		attested, ok := h.Value.(*AttestedDataCommitment)
		if !ok {
			panic(fmt.Errorf("invalid value type for SHARD_WORK_CONFIRMED selector: %T", h.Value))
		}
		return attested.ByteLength()
	case SHARD_WORK_PENDING:
		headers, ok := h.Value.(*PendingShardHeaders)
		if !ok {
			panic(fmt.Errorf("invalid value type for SHARD_WORK_PENDING selector: %T", h.Value))
		}
		return headers.ByteLength(spec)
	default:
		panic(errors.New("bad selector value"))
	}
}

func (h *ShardWorkStatus) FixedLength(spec *common.Spec) uint64 {
	return 0
}

func (h *ShardWorkStatus) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	if h.Selector == SHARD_WORK_UNCONFIRMED {
		return hFn.Union(h.Selector, nil)
	}
	attested, ok := h.Value.(*AttestedDataCommitment)
	if ok {
		return hFn.Union(h.Selector, attested)
	}
	headers, ok := h.Value.(*PendingShardHeaders)
	if !ok {
		return hFn.Union(h.Selector, spec.Wrap(headers))
	}
	return common.Root{}
}

func (h *ShardWorkStatus) View(spec *common.Spec) (*ShardWorkStatusView, error) {
	switch h.Selector {
	case SHARD_WORK_UNCONFIRMED:
		return AsShardWorkStatus(ShardWorkStatusType(spec).FromView(SHARD_WORK_UNCONFIRMED, nil))
	case SHARD_WORK_CONFIRMED:
		attested, ok := h.Value.(*AttestedDataCommitment)
		if !ok {
			return nil, fmt.Errorf("invalid value type for SHARD_WORK_CONFIRMED selector: %T", h.Value)
		}
		return AsShardWorkStatus(ShardWorkStatusType(spec).FromView(SHARD_WORK_CONFIRMED, attested.View()))
	case SHARD_WORK_PENDING:
		headers, ok := h.Value.(*PendingShardHeaders)
		if !ok {
			return nil, fmt.Errorf("invalid value type for SHARD_WORK_PENDING selector: %T", h.Value)
		}
		headersView, err := headers.View(spec)
		if err != nil {
			return nil, err
		}
		return AsShardWorkStatus(ShardWorkStatusType(spec).FromView(SHARD_WORK_PENDING, headersView))
	default:
		return nil, errors.New("bad selector value")
	}
}

type ShardWorkStatusView struct {
	*UnionView
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

func (h *ShardWork) View(spec *common.Spec) (*ShardWorkView, error) {
	statusView, err := h.Status.View(spec)
	if err != nil {
		return nil, err
	}
	return AsShardWork(ShardWorkType(spec).FromFields(statusView))
}

func ShardWorkType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("ShardWork", []FieldDef{
		{"status", ShardWorkStatusType(spec)},
	})
}

type ShardWorkView struct {
	*ContainerView
}

func AsShardWork(v View, err error) (*ShardWorkView, error) {
	c, err := AsContainer(v, err)
	return &ShardWorkView{c}, err
}

func (v *ShardWorkView) Status() (*ShardWorkStatusView, error) {
	c, err := AsUnion(v.Get(0))
	return &ShardWorkStatusView{c}, err
}
