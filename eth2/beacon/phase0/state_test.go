package phase0

import (
	"encoding/json"
	"testing"

	"github.com/protolambda/zrnt/eth2/configs"
	"gopkg.in/yaml.v3"
)

func TestWrapperJSONProxy(t *testing.T) {
	state := BeaconState{Slot: 123}
	spec := configs.Mainnet
	wrapped := spec.Wrap(&state)
	data, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	var other BeaconState
	wrappedOther := spec.Wrap(&other)
	err = json.Unmarshal(data, wrappedOther)
	if err != nil {
		t.Fatal(err)
	}
	if other.Slot != state.Slot {
		t.Fatalf("failed to marshal/unmarshal JSON roundtrip wrapped BeaconState: %d <> %d", other.Slot, state.Slot)
	}
}

func TestWrapperYAMLProxy(t *testing.T) {
	state := BeaconState{Slot: 123}
	spec := configs.Mainnet
	wrapped := spec.Wrap(&state)
	data, err := yaml.Marshal(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	var other BeaconState
	wrappedOther := spec.Wrap(&other)
	err = yaml.Unmarshal(data, wrappedOther)
	if err != nil {
		t.Fatal(err)
	}
	if other.Slot != state.Slot {
		t.Fatalf("failed to marshal/unmarshal JSON roundtrip wrapped BeaconState: %d <> %d", other.Slot, state.Slot)
	}
}
