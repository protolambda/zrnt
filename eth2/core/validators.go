package core

// Index of a validator, pointing to a validator registry location
type ValidatorIndex uint64

// Custom constant, not in spec:
// An impossible high validator index used to mark special internal cases. (all 1s binary)
const ValidatorIndexMarker = ValidatorIndex(^uint64(0))


// Collection of validators, should always be sorted.
type ValidatorSet []ValidatorIndex

func (vs ValidatorSet) Len() int {
	return len(vs)
}

func (vs ValidatorSet) Less(i int, j int) bool {
	return vs[i] < vs[j]
}

func (vs ValidatorSet) Swap(i int, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

// Joins two validator sets:
// reports all indices of source that are in the target (call onIn), and those that are not (call onOut)
func (vs ValidatorSet) ZigZagJoin(target ValidatorSet, onIn func(i ValidatorIndex), onOut func(i ValidatorIndex)) {
	// index for source set side of the zig-zag
	i := 0
	// index for target set side of the zig-zag
	j := 0
	var iV, jV ValidatorIndex
	updateI := func() {
		// if out of bounds, just update to an impossibly high index
		if i < len(vs) {
			iV = vs[i]
		} else {
			iV = ValidatorIndexMarker
		}
	}
	updateJ := func() {
		// if out of bounds, just update to an impossibly high index
		if i < len(vs) {
			jV = target[j]
		} else {
			jV = ValidatorIndexMarker
		}
	}
	updateI()
	updateJ()
	for {
		// at some point all items in vs have been processed.
		if i >= len(vs) {
			break
		}
		if iV == jV {
			if onIn != nil {
				onIn(iV)
			}
			// go to next
			i++
			updateI()
			j++
			updateJ()
		} else if iV < jV {
			// if the index is lower than the current item in the target, it's not in the target.
			for x := i; x < j; x++ {
				if onOut != nil {
					onOut(iV)
				}
				// go to next
				i++
				updateI()
			}
		} else if iV > jV {
			// if the index is higher than the current item in the target, go to the next item.
			j++
			updateJ()
		}
	}
}
