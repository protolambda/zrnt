package beacon

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const VersionType = Bytes4Type

// 32 bits, not strictly an integer, hence represented as 4 bytes
// (bytes not necessarily corresponding to versions)
type Version [4]byte

func (v *Version) Deserialize(dr *codec.DecodingReader) error {
	if v == nil {
		return errors.New("nil version")
	}
	_, err := dr.Read(v[:])
	return err
}

func (a Version) Serialize(w *codec.EncodingWriter) error {
	return w.Write(a[:])
}

func (a Version) ByteLength() uint64 {
	return 4
}

func (Version) FixedLength() uint64 {
	return 4
}

func (p Version) HashTreeRoot(_ tree.HashFn) (out Root) {
	copy(out[:], p[:])
	return
}

func (p Version) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p Version) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *Version) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil Version")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 8 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

func (v Version) ToUint32() uint32 {
	return uint32(v[0])<<24 | uint32(v[1])<<16 | uint32(v[2])<<8 | uint32(v[3])
}

func (v Version) View() SmallByteVecView {
	return v[:]
}

func AsVersion(v View, err error) (Version, error) {
	return AsBytes4(v, err)
}

// A digest of the current fork data
type ForkDigest [4]byte

func (p ForkDigest) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p ForkDigest) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *ForkDigest) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil Version")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 8 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

type ForkData struct {
	CurrentVersion        Version
	GenesisValidatorsRoot Root
}

func (v *ForkData) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&v.CurrentVersion, &v.GenesisValidatorsRoot)
}

func (*ForkData) FixedLength() uint64 {
	return 4 + 32
}

func (d *ForkData) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(d.CurrentVersion, d.GenesisValidatorsRoot)
}

func ComputeForkDataRoot(currentVersion Version, genesisValidatorsRoot Root) Root {
	data := ForkData{
		CurrentVersion:        currentVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}
	return data.HashTreeRoot(tree.GetHashFn())
}

func ComputeForkDigest(currentVersion Version, genesisValidatorsRoot Root) ForkDigest {
	var digest ForkDigest
	dataRoot := ComputeForkDataRoot(currentVersion, genesisValidatorsRoot)
	copy(digest[:], dataRoot[:4])
	return digest
}

type Fork struct {
	PreviousVersion Version
	CurrentVersion  Version
	Epoch           Epoch
}

func (b *Fork) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&b.PreviousVersion, &b.CurrentVersion, &b.Epoch)
}

func (a *Fork) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(a.PreviousVersion, a.CurrentVersion, a.Epoch)
}

func (a *Fork) ByteLength() uint64 {
	return ForkType.TypeByteLength()
}

func (a *Fork) FixedLength() uint64 {
	return ForkType.TypeByteLength()
}

func (a *Fork) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(a.PreviousVersion, a.CurrentVersion, a.Epoch)
}

func (f *Fork) View() *ForkView {
	res, _ := ForkType.FromFields(f.PreviousVersion.View(), f.CurrentVersion.View(), Uint64View(f.Epoch))
	return &ForkView{res}
}

var ForkType = ContainerType("Fork", []FieldDef{
	{"previous_version", VersionType},
	{"current_version", VersionType},
	{"epoch", EpochType}, // Epoch of latest fork
})

type ForkView struct{ *ContainerView }

func (f *ForkView) PreviousVersion() (Version, error) {
	return AsVersion(f.Get(0))
}

func (f *ForkView) CurrentVersion() (Version, error) {
	return AsVersion(f.Get(1))
}

func (f *ForkView) Epoch() (Epoch, error) {
	return AsEpoch(f.Get(2))
}

func (f *ForkView) Raw() (*Fork, error) {
	prev, err := f.PreviousVersion()
	if err != nil {
		return nil, err
	}
	curr, err := f.CurrentVersion()
	if err != nil {
		return nil, err
	}
	ep, err := f.Epoch()
	if err != nil {
		return nil, err
	}
	return &Fork{
		PreviousVersion: prev,
		CurrentVersion:  curr,
		Epoch:           ep,
	}, nil
}

func AsFork(v View, err error) (*ForkView, error) {
	c, err := AsContainer(v, err)
	return &ForkView{c}, err
}

// Return the signature domain (fork version concatenated with domain type) of a message.
func (state *BeaconStateView) GetDomain(dom BLSDomainType, messageEpoch Epoch) (BLSDomain, error) {
	forkView, err := state.Fork()
	if err != nil {
		return BLSDomain{}, err
	}
	fork, err := forkView.Raw()
	if err != nil {
		return BLSDomain{}, err
	}
	var v Version
	if messageEpoch < fork.Epoch {
		v = fork.PreviousVersion
	} else {
		v = fork.CurrentVersion
	}
	genesisValRoot, err := state.GenesisValidatorsRoot()
	if err != nil {
		return BLSDomain{}, err
	}
	// combine fork version with domain type.
	return ComputeDomain(dom, v, genesisValRoot), nil
}
