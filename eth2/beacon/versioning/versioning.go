package versioning

import (
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)

type GenesisTimeProp TimestampReadProp

func (p GenesisTimeProp) GenesisTime() (Timestamp, error) {
	return (TimestampReadProp)(p).Timestamp()
}

type GenesisValidatorsRootProp RootReadProp

func (p GenesisValidatorsRootProp) GenesisValidatorsRoot() (Root, error) {
	return (RootReadProp)(p).Root()
}

var ForkType = &ContainerType{
	{"previous_version", VersionType},
	{"current_version", VersionType},
	{"epoch", EpochType}, // Epoch of latest fork
}

type Fork struct { *ContainerView }

func (f *Fork) PreviousVersion() (Version, error) {
	return VersionProp(PropReader(f, 0)).Version()
}

func (f *Fork) CurrentVersion() (Version, error) {
	return VersionProp(PropReader(f, 1)).Version()
}

func (f *Fork) ForkEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(f, 2)).Epoch()
}

type ForkProp ReadPropFn

func (p ForkProp) Fork() (*Fork, error) {
	if v, err := p(); err != nil {
		return nil, err
	} else if f, ok := v.(*Fork); !ok {
		return nil, fmt.Errorf("not a fork view: %v", v)
	} else {
		return f, nil
	}
}


type CurrentSlotReadProp SlotReadProp

func (p CurrentSlotReadProp) CurrentSlot() (Slot, error) {
	return (SlotReadProp)(p).Slot()
}

// Get current epoch
func (p CurrentSlotReadProp) CurrentEpoch() (Epoch, error) {
	if slot, err := p.CurrentSlot(); err != nil {
		return 0, nil
	} else {
		return slot.ToEpoch(), nil
	}
}

// Return previous epoch.
func (p CurrentSlotReadProp) PreviousEpoch() (Epoch, error) {
	if epoch, err := p.CurrentEpoch(); err != nil {
		return 0, nil
	} else {
		return epoch.Previous(), nil
	}
}

type CurrentSlotMutProp struct {
	CurrentSlotReadProp
	SlotWriteProp
}

func (p CurrentSlotMutProp) IncrementSlot() error {
	v, err := p.CurrentSlot()
	if err != nil {
		return err
	}
	return p.SetSlot(v + 1)
}

type DomainFeature struct {
	Meta interface {
		meta.GenesisValidatorsRoot
		meta.Forking
	}
}

// Return the signature domain (fork version concatenated with domain type) of a message.
func (f *DomainFeature) GetDomain(dom BLSDomainType, messageEpoch Epoch) (BLSDomain, error) {
	forkEpoch, err := f.Meta.ForkEpoch()
	if err != nil {
		return BLSDomain{}, err
	}
	var v Version
	if messageEpoch < forkEpoch {
		v, err = f.Meta.PreviousVersion()
	} else {
		v, err = f.Meta.CurrentVersion()
	}
	if err != nil {
		return BLSDomain{}, err
	}
	genesisValRoot, err := f.Meta.GenesisValidatorsRoot()
	if err != nil {
		return BLSDomain{}, err
	}
	// combine fork version with domain type.
	return ComputeDomain(dom, v, genesisValRoot), nil
}

type VersioningProps struct {
	GenesisTimeProp
	GenesisValidatorsRootProp
	CurrentSlotMutProp
	ForkProp
}
