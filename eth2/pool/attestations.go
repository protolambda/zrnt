package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/tree"
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
	Participants phase0.AttestationBits
	Sig          common.BLSSignature
}

type MinAggregates struct {
	Aggregates []Aggregate
	// The OR of all bitfields contained in Aggregates list, to easily filter out subsets
	Participants phase0.AttestationBits
	// Things already covered by the sum of the above aggregates, but maybe useful later. Keep a limited number of these.
	Extra []Aggregate
}

type AttestationPool struct {
	sync.RWMutex
	spec *common.Spec
	// att data root -> (data contents, committee indices)
	datas map[common.Root]*IndexedAttData
	// (validator, epoch) -> individual attestation
	individual map[Assignment]*AttRef
	// att data root -> minimum representation of everything attesting to it
	aggregate map[common.Root]*MinAggregates
	// (validator, epoch) -> att data root.
	// When a validator participates in an aggregate, remember which attestation data the validator is attesting too.
	// This helps filter duplicate aggregate attestations:
	// if all aggregate participants already voted, it can be ignored (and maybe slashed if bad double votes).
	aggPerValidator map[Assignment]common.Root
	// Keep some extra data around, which is already covered by larger aggregates, to try and pack better results.
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

func (ap *AttestationPool) AddAttestation(ctx context.Context, att *phase0.Attestation, committee common.CommitteeIndices) error {
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

	// aggregates: don't store more than we have to.
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

			// remember the participants attested this epoch
			key := Assignment{Index: 0, Epoch: att.Data.Target.Epoch}
			for i, vi := range committee {
				if att.AggregationBits.GetBit(uint64(i)) {
					key.Index = vi
					ap.aggPerValidator[key] = dataRoot
				}
			}
			return nil
		}
	} else {
		hasNewAttester := false
		key := Assignment{Index: 0, Epoch: att.Data.Target.Epoch}
		// check if we have not seen any of the participants attest this epoch yet
		for i, vi := range committee {
			if att.AggregationBits.GetBit(uint64(i)) {
				key.Index = vi
				if _, ok := ap.aggPerValidator[key]; !ok {
					hasNewAttester = true
					ap.aggPerValidator[key] = dataRoot
				}
			}
		}
		if hasNewAttester {
			ap.aggregate[dataRoot] = &MinAggregates{
				Aggregates: []Aggregate{{Participants: att.AggregationBits, Sig: att.Signature}},
				// copy, we mutate this bitfield later, while still using the original (stored in above array)
				Participants: att.AggregationBits.Copy(),
			}
		} else {
			return fmt.Errorf("ignoring new attestation for different data:" +
				"all participants voted for other data this epoch already, whole attestation is likely slashable")
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

// Prune pool based on current epoch, attestations which cannot be included anymore will get pruned.
func (ap *AttestationPool) Prune(epoch common.Epoch) {
	min := epoch.Previous()
	for k, v := range ap.datas {
		if v.Data.Target.Epoch < min {
			delete(ap.datas, k)
			delete(ap.aggregate, k)
		}
	}
	for k := range ap.individual {
		if k.Epoch < min {
			delete(ap.individual, k)
		}
	}
	for k := range ap.aggPerValidator {
		if k.Epoch < min {
			delete(ap.aggPerValidator, k)
		}
	}
}

// Approximation of the optimal attestation packing.
// Attestations must match source, get prioritized if the target is correct, and more if the head is correct.
// Attestations may not be included if they already are (checked via included func).
// Maximum attestation output and packing-time constraints apply.
func (ap *AttestationPool) Packing(ctx context.Context,
	source common.Checkpoint, target common.Checkpoint,
	headRoot common.Root, headSlot common.Slot,
	maxCount uint64, maxTime time.Duration,
	included func(epoch common.Epoch, index common.ValidatorIndex)) ([]phase0.Attestation, error) {

	// TODO find best attestations to pack for profit.
	return nil, nil
}
