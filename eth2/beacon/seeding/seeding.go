package seeding

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
)

type SeedFeature struct {
	Meta interface {
		meta.Randomness
	}
}

// Generate a seed for the given epoch
func (f *SeedFeature) GetSeed(epoch Epoch, domainType BLSDomainType) Root {
	buf := make([]byte, 4+8+32)

	// domain type
	copy(buf[0:4], domainType[:])

	// epoch
	binary.LittleEndian.PutUint64(buf[4:4+8], uint64(epoch))

	// Avoid underflow
	mix := f.Meta.GetRandomMix(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD - 1)
	copy(buf[4+8:], mix[:])

	return Hash(buf)
}
