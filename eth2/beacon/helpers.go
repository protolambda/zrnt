package beacon

// Return the validators from src that are inside the targets list, and those that are outside.
func FindInAndOutValidators(src []ValidatorIndex, targets []ValidatorIndex) ([]ValidatorIndex, []ValidatorIndex) {

	indexMap := make(map[ValidatorIndex]bool)
	for _, vIndex := range src {
		indexMap[vIndex] = true
	}
	// the good indices from source will be put on the left side, the bad on the right.
	res := make([]ValidatorIndex, len(src), len(src))
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
