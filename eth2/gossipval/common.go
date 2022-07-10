package gossipval

import (
	"fmt"
	"time"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// CheckSlotSpan checks if the slot is within the span of slots, with MAXIMUM_GOSSIP_CLOCK_DISPARITY margin in time.
func CheckSlotSpan(slotAfter func(delta time.Duration) common.Slot, slot common.Slot, span common.Slot) error {
	if slot+span < slot {
		return fmt.Errorf("slot overflow: %d", slot)
	}
	// check minimum, with account for clock disparity
	if minSlot := slotAfter(-MAXIMUM_GOSSIP_CLOCK_DISPARITY); slot+span < minSlot {
		return fmt.Errorf("slot %d is too old, minimum slot is %d", slot, minSlot)
	}
	// check maximum, with account for clock disparity
	if maxSlot := slotAfter(MAXIMUM_GOSSIP_CLOCK_DISPARITY); slot > maxSlot {
		return fmt.Errorf("slot %d is too new, maximum slot is %d", slot, maxSlot)
	}
	return nil
}
