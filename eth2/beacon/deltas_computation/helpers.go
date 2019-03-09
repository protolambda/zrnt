package deltas_computation

import "github.com/protolambda/zrnt/eth2/beacon"

// Return the validators from src that are inside the targets list, and those that are outside.
func FindInAndOutValidators(src []beacon.ValidatorIndex, targets []beacon.ValidatorIndex) (
	[]beacon.ValidatorIndex, []beacon.ValidatorIndex) {

	indexMap := make(map[beacon.ValidatorIndex]bool)
	for _, vIndex := range src {
		indexMap[vIndex] = true
	}
	// the good indices from source will be put on the left side, the bad on the right.
	res := make([]beacon.ValidatorIndex, len(src), len(src))
	i := 0
	j := len(res) - 1
	for _, e := range targets {
		if indexMap[e] {
			res[i] = e
			i++
		} else {
			res[j] = e
			j--
		}
	}

	// return as two slices, the inside validators, and the outside validators.
	return res[:i], res[i:]
}
