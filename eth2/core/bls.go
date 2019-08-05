package core

type BLSPubkey [48]byte

type BLSSignature [96]byte

// Mixed into a BLS domain to define its type
type BLSDomainType uint32

// BLS domain (64 bits): fork version (32 bits) concatenated with BLS domain type (32 bits)
type BLSDomain uint64

func ComputeDomain(domainType BLSDomainType, forkVersion Version) BLSDomain {
	return BLSDomain((uint64(domainType) << 32) | uint64(forkVersion.ToUint32()))
}
