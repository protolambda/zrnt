package ssz

import (
	"encoding/binary"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/util/merkle"
	"reflect"
)

// constructs a merkle_root of the given struct (panics if it's not a struct, or a pointer to one),
// but ignores any fields that are tagged with `ssz:"signature"`
func Signed_root(input interface{}) eth2.Root {
	subRoots := make([]eth2.Bytes32, 0)
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic("cannot get partial root for signing, input is not a struct")
	}
	vType := v.Type()
	for i, fields := 0, v.NumField(); i < fields; i++ {
		// ignore all fields with a signatures
		if tag, ok := vType.Field(i).Tag.Lookup("ssz"); ok && tag == "signature" {
			break
		}
		subRoots = append(subRoots, eth2.Bytes32(sszHashTreeRoot(v.Field(i))))
	}
	return merkle.Merkle_root(subRoots)
}

func SSZEncode(input interface{}) []byte {
	out := make([]byte, 0)
	sszSerialize(reflect.ValueOf(input), &out)
	return out
}

func withSize(dst *[]byte, size uint64) (start uint64, end uint64) {
	// if capacity is too low, extend it.
	start, end = uint64(len(*dst)), uint64(len(*dst))+size
	if uint64(cap(*dst)) < end {
		res := make([]byte, end, end*2)
		copy(res[0:start], *dst)
		*dst = res
	}
	*dst = (*dst)[:end]
	return start, end
}

func sszSerialize(v reflect.Value, dst *[]byte) (encodedLen uint32) {
	switch v.Kind() {
	case reflect.Ptr:
		return sszSerialize(v.Elem(), dst)
	case reflect.Uint8: // "uintN"
		s, _ := withSize(dst, 1)
		(*dst)[s] = byte(v.Uint())
		return 1
		// Commented, not really used in spec.
		//case reflect.Uint32: // "uintN"
		//	s, e := withSize(dst, 4)
		//	binary.LittleEndian.PutUint32((*dst)[s:e], uint32(v.Uint()))
		//return 4
	case reflect.Uint64: // "uintN"
		s, e := withSize(dst, 8)
		binary.LittleEndian.PutUint64((*dst)[s:e], uint64(v.Uint()))
		return 8
	case reflect.Bool: // "bool"
		s, _ := withSize(dst, 1)
		if v.Bool() {
			(*dst)[s] = 1
		} else {
			(*dst)[s] = 0
		}
		return 1
	case reflect.Array: // "tuple"
		// TODO: We're ignoring that arrays with variable sized items (eg. slices) are a thing in Go. Don't use them.
		// Possible workarounds for this: (i) check sizes before encoding. (ii) panic if serializedSize is irregular.
		// Special fields (e.g. "Root", "Bytes32" will just be packed as packed arrays, which is fine, little-endian!)
		for i, size := 0, v.Len(); i < size; i++ {
			serializedSize := sszSerialize(v.Index(i), dst)
			encodedLen += serializedSize
		}
		return encodedLen
	case reflect.Slice: // "list"
		for i, size := 0, v.Len(); i < size; i++ {
			// allocate size prefix: BYTES_PER_LENGTH_PREFIX
			s, e := withSize(dst, 4)
			serializedSize := sszSerialize(v.Index(i), dst)
			binary.LittleEndian.PutUint32((*dst)[s:e], serializedSize)
			encodedLen += 4 + serializedSize
		}
		return encodedLen
	case reflect.Struct: // "container"
		for i, size := 0, v.NumField(); i < size; i++ {
			// allocate size prefix: BYTES_PER_LENGTH_PREFIX
			s, e := withSize(dst, 4)
			serializedSize := sszSerialize(v.Field(i), dst)
			binary.LittleEndian.PutUint32((*dst)[s:e], serializedSize)
			encodedLen += 4 + serializedSize
		}
		return encodedLen
	default:
		panic("ssz encoding: unsupported value kind: " + v.Kind().String())
	}
}

func Hash_tree_root(input interface{}) eth2.Root {
	return sszHashTreeRoot(reflect.ValueOf(input))
}

// TODO: see specs #679, comment.
// Implementation here simply assumes fixed-length arrays only have elements of fixed-length.

// Compute hash tree root for a value
func sszHashTreeRoot(v reflect.Value) eth2.Root {
	switch v.Kind() {
	case reflect.Ptr:
		return sszHashTreeRoot(v.Elem())
	// "basic object or a tuple of basic objects"
	case reflect.Uint8, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Array:
		return merkle.Merkle_root(sszPack(v))
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		// "list of basic objects"
		case reflect.Uint8, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Array:
			packedData := sszPack(v)
			root := merkle.Merkle_root(packedData)
			return sszMixInLength(root, uint64(v.Len()))
		// Interpretation: list of composite / var-size (i.e. the non-basic) objects
		default:
			data := make([]eth2.Bytes32, v.Len())
			for i := 0; i < v.Len(); i++ {
				data[i] = eth2.Bytes32(sszHashTreeRoot(v.Index(i)))
			}
			return sszMixInLength(merkle.Merkle_root(data), uint64(v.Len()))
		}
	// Interpretation: container, similar to list of complex objects, but without length prefix.
	case reflect.Struct:
		data := make([]eth2.Bytes32, v.NumField())
		for i, length := 0, v.NumField(); i < length; i++ {
			data[i] = eth2.Bytes32(sszHashTreeRoot(v.Field(i)))
		}
		return merkle.Merkle_root(data)
	default:
		panic("tree-hash: unsupported value kind: " + v.Kind().String())
	}
}

func sszPack(input reflect.Value) []eth2.Bytes32 {
	serialized := make([]byte, 0)
	switch input.Kind() {
	case reflect.Array, reflect.Slice:
		for i, length := 0, input.Len(); i < length; i++ {
			sszSerialize(input.Index(i), &serialized)
		}
	case reflect.Struct:
		for i, length := 0, input.NumField(); i < length; i++ {
			sszSerialize(input.Field(i), &serialized)
		}
	default:
		sszSerialize(input, &serialized)
	}
	// floored: handle all normal chunks first
	flooredChunkCount := len(serialized) / 32
	// ceiled: include any partial chunk at end as full chunk (with padding)
	out := make([]eth2.Bytes32, (len(serialized)+31)/32)
	for i := 0; i < flooredChunkCount; i++ {
		copy(out[i][:], serialized[i<<5:(i+1)<<5])
	}
	// if there is a partial chunk at the end, handle it as a special case:
	if len(serialized)&31 != 0 {
		copy(out[flooredChunkCount][:len(serialized)&0x1F], serialized[flooredChunkCount<<5:])
	}
	return out
}

func sszMixInLength(data eth2.Root, length uint64) eth2.Root {
	lengthInput := eth2.Bytes32{}
	binary.LittleEndian.PutUint64(lengthInput[:], length)
	return merkle.Merkle_root([]eth2.Bytes32{eth2.Bytes32(data), lengthInput})
}
