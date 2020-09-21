package beacon

import (
	"errors"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var DepositRootsType = ComplexListType(RootType, 1<<DEPOSIT_CONTRACT_TREE_DEPTH)

type DepositRootsView struct {
	*ComplexListView
}

func AsDepositRootsView(v View, err error) (*DepositRootsView, error) {
	c, err := AsComplexList(v, err)
	return &DepositRootsView{c}, err
}

func NewDepositRootsView() *DepositRootsView {
	return &DepositRootsView{DepositRootsType.New()}
}

func (spec *Spec) GenesisFromEth1(eth1BlockHash Root, time Timestamp, deps []Deposit, ignoreSignaturesAndProofs bool) (*BeaconStateView, *EpochsContext, error) {
	state := spec.NewBeaconStateView()
	if err := state.SetGenesisTime(time + spec.GENESIS_DELAY); err != nil {
		return nil, nil, err
	}
	if err := state.SetFork(Fork{
		PreviousVersion: spec.GENESIS_FORK_VERSION,
		CurrentVersion:  spec.GENESIS_FORK_VERSION,
		Epoch:           GENESIS_EPOCH,
	}); err != nil {
		return nil, nil, err
	}
	eth1Dat := Eth1Data{
		DepositRoot:  Root{}, // incrementally overwritten during deposit processing
		DepositCount: DepositIndex(len(deps)),
		BlockHash:    eth1BlockHash,
	}
	if err := state.SetEth1Data(eth1Dat.View()); err != nil {
		return nil, nil, err
	}
	emptyBody := BeaconBlockBody{}
	latestHeader := BeaconBlockHeader{
		BodyRoot: emptyBody.HashTreeRoot(spec, tree.GetHashFn()),
	}
	if err := state.SetLatestBlockHeader(latestHeader.View()); err != nil {
		return nil, nil, err
	}
	// Seed RANDAO with Eth1 entropy
	randaoMixes, err := spec.SeedRandao(eth1BlockHash)
	if err != nil {
		return nil, nil, err
	}
	if err := state.SetRandaoMixes(randaoMixes); err != nil {
		return nil, nil, err
	}

	pc, err := NewPubkeyCache(state)
	if err != nil {
		return nil, nil, err
	}
	// Create mostly empty epochs context. Just need the pubkey cache first
	epc := &EpochsContext{
		Spec:        spec,
		PubkeyCache: pc,
	}

	depRootsView := NewDepositRootsView()

	hFn := tree.GetHashFn()
	updateDepTreeRoot := func() error {
		eth1DatView, err := state.Eth1Data()
		if err != nil {
			return err
		}
		depTreeRoot := depRootsView.HashTreeRoot(hFn)
		return eth1DatView.SetDepositRoot(depTreeRoot)
	}
	// Process deposits
	for i := range deps {
		depRoot := RootView(deps[i].Data.HashTreeRoot(tree.GetHashFn()))
		if err := depRootsView.Append(&depRoot); err != nil {
			return nil, nil, err
		}
		if err := updateDepTreeRoot(); err != nil {
			return nil, nil, err
		}
		// in the rare case someone tries to create a genesis block using invalid data, error.
		if err := spec.ProcessDeposit(epc, state, &deps[i], ignoreSignaturesAndProofs); err != nil {
			return nil, nil, err
		}
	}
	if err := updateDepTreeRoot(); err != nil {
		return nil, nil, err
	}
	vals, err := state.Validators()
	if err != nil {
		return nil, nil, err
	}
	valCount, err := vals.Length()
	if err != nil {
		return nil, nil, err
	}
	if Slot(valCount) < spec.SLOTS_PER_EPOCH {
		return nil, nil, errors.New("not enough validators to init full featured BeaconState")
	}

	// Process activations
	for i := uint64(0); i < valCount; i++ {
		val, err := AsValidator(vals.Get(i))
		if err != nil {
			return nil, nil, err
		}
		vEff, err := val.EffectiveBalance()
		if err != nil {
			return nil, nil, err
		}
		if vEff == spec.MAX_EFFECTIVE_BALANCE {
			if err := val.SetActivationEligibilityEpoch(GENESIS_EPOCH); err != nil {
				return nil, nil, err
			}
			if err := val.SetActivationEpoch(GENESIS_EPOCH); err != nil {
				return nil, nil, err
			}
		}
	}
	if err := state.SetGenesisValidatorsRoot(vals.HashTreeRoot(hFn)); err != nil {
		return nil, nil, err
	}
	// Complete computation of epc
	if err := epc.LoadShuffling(state); err != nil {
		return nil, nil, err
	}
	if err := epc.LoadProposers(state); err != nil {
		return nil, nil, err
	}
	return state, epc, nil
}

func (spec *Spec) IsValidGenesisState(state *BeaconStateView) (bool, error) {
	genTime, err := state.GenesisTime()
	if err != nil {
		return false, err
	}
	if genTime < spec.MIN_GENESIS_TIME {
		return false, nil
	}

	// outside of genesis we have this precomputed at all times. Just compute it manually this time.
	activeCount := uint64(0)
	{
		validators, err := state.Validators()
		if err != nil {
			return false, err
		}
		valIter := validators.ReadonlyIter()
		for {
			valContainer, ok, err := valIter.Next()
			if err != nil {
				return false, err
			}
			if !ok {
				break
			}
			val, err := AsValidator(valContainer, nil)
			if err != nil {
				return false, err
			}
			if active, err := spec.IsActive(val, GENESIS_EPOCH); err != nil {
				return false, err
			} else if active {
				activeCount += 1
			}
		}
	}
	return activeCount >= spec.MIN_GENESIS_ACTIVE_VALIDATOR_COUNT, nil
}
