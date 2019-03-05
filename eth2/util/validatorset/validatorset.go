package validatorset

import (
	"go-beacon-transition/eth2"
)

type ValidatorIndexSet []eth2.ValidatorIndex

// returns a copy without the given indices
func (vs ValidatorIndexSet) Minus(removed ValidatorIndexSet) ValidatorIndexSet {
	res := vs.Copy()
	res.RemoveAll(removed)
	return res
}

func (vs *ValidatorIndexSet) RemoveAll(removed ValidatorIndexSet) {
	// TODO possible optimization: sort both, and zip-join: iterate both in parallel, marking matches.
	for i, a := range *vs {
		for _, b := range removed {
			if a == b {
				(*vs)[i] = eth2.ValidatorIndexMarker
				break
			}
		}
	}
	// remove all marked indices
	for i := 0; i < len(*vs); {
		if (*vs)[i] == eth2.ValidatorIndexMarker {
			// replace with last, and cut out last
			last := len(*vs) - 1
			(*vs)[i] = (*vs)[last]
			*vs = (*vs)[:last]
		} else {
			i++
		}
	}
}

func (vs ValidatorIndexSet) Copy() ValidatorIndexSet {
	res := make([]eth2.ValidatorIndex, len(vs), len(vs))
	copy(res, vs)
	return res
}
