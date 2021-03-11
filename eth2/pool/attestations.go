package pool

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/tree"
	"sync"
	"time"
)

type Assignment struct {
	Index common.ValidatorIndex
	Epoch common.Epoch
}

type IndexedAttData struct {
	Data      phase0.AttestationData
	Committee common.CommitteeIndices
}

type AttRef struct {
	DataRoot common.Root
	Sig      common.BLSSignature
}

type Aggregate struct {
	Participants common.CommitteeBits
	Sig          common.BLSSignature
}

type MinAggregates struct {
	Aggregates []Aggregate
	// The OR of all bitfields contained in Aggregates list, to easily filter out subsets
	Participants common.CommitteeBits
	// Things already covered by the sum of the above aggregates, but maybe useful later. Keep a limited number of these.
	Extra []Aggregate
}

type AttestationPool struct {
	sync.RWMutex
	spec               *common.Spec
	datas              map[common.Root]*IndexedAttData
	individual         map[Assignment]*AttRef
	aggregate          map[common.Root]*MinAggregates
	maxExtraAggregates uint64
}

func NewAttestationPool(spec *common.Spec) *AttestationPool {
	return &AttestationPool{
		spec:               spec,
		datas:              make(map[common.Root]*IndexedAttData),
		individual:         make(map[Assignment]*AttRef),
		aggregate:          make(map[common.Root]*MinAggregates),
		maxExtraAggregates: 10, // TODO: worth tuning
	}
}

func (ap *AttestationPool) AddAttestation(att *phase0.Attestation, committee common.CommitteeIndices) error {
	ap.Lock()
	defer ap.Unlock()

	count := att.AggregationBits.OnesCount()
	if count == 0 {
		return errors.New("empty attestations are not allowed")
	}

	// store data and committee, so we won't have to inevitably fetch the info from a state or cache later.
	dataRoot := att.Data.HashTreeRoot(tree.GetHashFn())
	if _, ok := ap.datas[dataRoot]; !ok {
		ap.datas[dataRoot] = &IndexedAttData{
			Data:      att.Data,
			Committee: committee,
		}
	}

	// unaggregated attestation: track separately. For efficiency and easy aggregation.
	if count == 1 {
		val, err := att.AggregationBits.SingleParticipant(committee)
		if err != nil { // e.g. the bitfield length doesn't match the committee.
			return fmt.Errorf("could not get attestation participant from bitfield and committee combi: %v", err)
		}
		key := Assignment{Index: val, Epoch: att.Data.Target.Epoch}
		if existing, ok := ap.individual[key]; ok {
			if existing.DataRoot != dataRoot {
				// double votes are slashable bad behavior. We mark it as a bad attestation.
				return fmt.Errorf("double vote by: %d, epoch %d, data root: %s", key.Index, key.Epoch, dataRoot)
			} else {
				// already have this exact attestation in the pool
				return nil
			}
		}
		ap.individual[key] = &AttRef{DataRoot: dataRoot, Sig: att.Signature}
		return nil
	}

	// aggregates: don't store than we have to.
	// Sometimes we find some different ones, keep those, every attester counts.
	// No aggregation yet, we can put together the best version later.
	if existing, ok := ap.aggregate[dataRoot]; ok {
		if covers, err := existing.Participants.Covers(att.AggregationBits); err != nil {
			return fmt.Errorf("could not compare aggregation bitfields: %v", err)
		} else if covers {
			// New attestation doesn't add any new info,
			// but if it packs better than something we had before, we should keep it for better performance.
			// To avoid spam / DoS, we only keep a limited number of these
			if uint64(len(existing.Extra)) < ap.maxExtraAggregates {
				existing.Extra = append(existing.Extra,
					Aggregate{Participants: att.AggregationBits, Sig: att.Signature})
			}
			return nil
		} else {
			// this aggregate adds additional participants compared to the total we had before, keep it!
			existing.Aggregates = append(existing.Aggregates,
				Aggregate{Participants: att.AggregationBits, Sig: att.Signature})
			return nil
		}
	} else {
		ap.aggregate[dataRoot] = &MinAggregates{
			Aggregates: []Aggregate{{Participants: att.AggregationBits, Sig: att.Signature}},
			// copy, we mutate this bitfield later, while still using the original (stored in above array)
			Participants: att.AggregationBits.Copy(),
		}
		return nil
	}
}

type attSearch struct {
	slot *common.Slot
	comm *common.CommitteeIndex
}

type AttSearchOption func(a *attSearch)

func WithSlot(slot common.Slot) AttSearchOption {
	return func(a *attSearch) {
		a.slot = &slot
	}
}

func WithCommittee(index common.CommitteeIndex) AttSearchOption {
	return func(a *attSearch) {
		a.comm = &index
	}
}

func (ap *AttestationPool) Search(opts ...AttSearchOption) (out []*phase0.Attestation) {
	var conf attSearch
	for _, opt := range opts {
		opt(&conf)
	}
	for k, d := range ap.datas {
		if conf.slot != nil && d.Data.Slot != *conf.slot {
			continue
		}
		if conf.comm != nil && d.Data.Index != *conf.comm {
			continue
		}
		agg := ap.aggregate[k]
		for _, a := range agg.Aggregates {
			out = append(out, &phase0.Attestation{AggregationBits: a.Participants, Data: d.Data, Signature: a.Sig})
		}
		// TODO: could add individual attestations
	}
	return out
}

func (ap *AttestationPool) Prune() {
	// TODO
}

func (ap *AttestationPool) Packing(ctx context.Context, source common.Checkpoint, target common.Checkpoint,
	maxCount uint64, maxTime time.Duration) ([]phase0.Attestation, error) {

	// TODO find best attestations to pack for profit.
	return nil, nil
}
