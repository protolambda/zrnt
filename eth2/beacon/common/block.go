package common

import (
	"bytes"
	"github.com/protolambda/zrnt/eth2/util/bls"
)

type BeaconBlockEnvelope struct {
	ForkDigest ForkDigest

	// Block header details
	Slot          Slot
	ProposerIndex ValidatorIndex
	ParentRoot    Root
	StateRoot     Root
	// TODO: maybe add body-root?

	// Fork-specific block container
	SignedBlock SpecObj

	// Cached block root (hash-tree-root of Message)
	BlockRoot Root
	// Block signature
	Signature BLSSignature
}

func (b *BeaconBlockEnvelope) VerifySignature(spec *Spec, genesisValidatorsRoot Root, proposer ValidatorIndex, pub *CachedPubkey) bool {
	version := spec.ForkVersion(b.Slot)
	return b.VerifySignatureVersioned(spec, version, genesisValidatorsRoot, proposer, pub)
}

// deprecated: to verify with explicit version
func (b *BeaconBlockEnvelope) VerifySignatureVersioned(spec *Spec, version Version, genesisValidatorsRoot Root, proposer ValidatorIndex, pub *CachedPubkey) bool {
	if b.ProposerIndex != proposer {
		return false
	}
	forkRoot := ComputeForkDataRoot(version, genesisValidatorsRoot)
	// Sanity check fork digest
	if !bytes.Equal(forkRoot[0:4], b.ForkDigest[:]) {
		return false
	}
	dom := ComputeDomain(spec.DOMAIN_BEACON_PROPOSER, version, genesisValidatorsRoot)
	return bls.Verify(pub, ComputeSigningRoot(b.BlockRoot, dom), b.Signature)
}

type EnvelopeBuilder interface {
	Envelope(spec *Spec, digest ForkDigest) *BeaconBlockEnvelope
}
