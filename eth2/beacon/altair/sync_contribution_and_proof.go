package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type ContributionAndProof struct {
	AggregatorIndex common.ValidatorIndex     `yaml:"aggregator_index" json:"aggregator_index"`
	Contribution    SyncCommitteeContribution `yaml:"contribution" json:"contribution"`
	SelectionProof  common.BLSSignature       `yaml:"selection_proof" json:"selection_proof"`
}

func ContributionAndProofType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("ContributionAndProof", []FieldDef{
		{"aggregator_index", common.ValidatorIndexType},
		{"contribution", SyncCommitteeContributionType(spec)},
		{"selection_proof", common.BLSSignatureType},
	})
}

func (agg *ContributionAndProof) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
		&agg.AggregatorIndex,
		spec.Wrap(&agg.Contribution),
		&agg.SelectionProof,
	)
}

func (agg *ContributionAndProof) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(
		&agg.AggregatorIndex,
		spec.Wrap(&agg.Contribution),
		&agg.SelectionProof,
	)
}

func (agg *ContributionAndProof) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&agg.AggregatorIndex,
		spec.Wrap(&agg.Contribution),
		&agg.SelectionProof,
	)
}

func (agg *ContributionAndProof) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&agg.AggregatorIndex,
		spec.Wrap(&agg.Contribution),
		&agg.SelectionProof,
	)
}

func (agg *ContributionAndProof) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&agg.AggregatorIndex,
		spec.Wrap(&agg.Contribution),
		&agg.SelectionProof,
	)
}

type ContributionAndProofView struct {
	*ContainerView
}

func AsContributionAndProof(v View, err error) (*ContributionAndProofView, error) {
	c, err := AsContainer(v, err)
	return &ContributionAndProofView{c}, err
}

func SignedContributionAndProofType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SignedContributionAndProof", []FieldDef{
		{"message", ContributionAndProofType(spec)},
		{"signature", common.BLSSignatureType},
	})
}

type SignedContributionAndProof struct {
	Message   ContributionAndProof `yaml:"message" json:"message"`
	Signature common.BLSSignature  `yaml:"signature" json:"signature"`
}

func (b *SignedContributionAndProof) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedContributionAndProof) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedContributionAndProof) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&b.Message), &b.Signature)
}

func (a *SignedContributionAndProof) FixedLength(*common.Spec) uint64 {
	return 0
}

func (b *SignedContributionAndProof) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&b.Message), b.Signature)
}

type SignedContributionAndProofView struct {
	*ContainerView
}

func AsSignedContributionAndProof(v View, err error) (*SignedContributionAndProofView, error) {
	c, err := AsContainer(v, err)
	return &SignedContributionAndProofView{c}, err
}
