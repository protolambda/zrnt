package compact

import . "github.com/protolambda/zrnt/eth2/core"

type CompactValidator uint64

func MakeCompactValidator(index ValidatorIndex, slashed bool, effectiveBalance Gwei) CompactValidator {
	compactData := CompactValidator(index) << 16
	if slashed {
		compactData |= 1 << 15
	}
	compactData |= CompactValidator(effectiveBalance / EFFECTIVE_BALANCE_INCREMENT)
	return compactData
}

func (cv CompactValidator) Index() ValidatorIndex {
	return ValidatorIndex(cv >> 16)
}

func (cv CompactValidator) Slashed() bool {
	return ((cv >> 15) & 1) == 1
}

func (cv CompactValidator) EffectiveBalance() Gwei {
	return Gwei(cv & ((1 << 15) - 1))
}
