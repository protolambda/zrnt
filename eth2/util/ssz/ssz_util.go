package ssz

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/tags"
	"reflect"
)

const SSZ_TAG = "ssz"
const OMIT_FLAG = "omit"

// Note: when this is changed,
//  don't forget to change the PutUint32 calls that make put the length in this allocated space.
const BYTES_PER_LENGTH_OFFSET = 4

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

func getSerializeCache(v reflect.Value) *SerializeCache {
	prov, ok := v.Interface().(SSZSerializeCacheProvider)
	if !ok {
		// not a cache provider, no cache available
		return nil
	} else {
		return prov.GetSerializeCache()
	}
}

func sszSerializeMaybeCached(v reflect.Value, dst *[]byte) {
	if cache := getSerializeCache(v); cache != nil {
		// Cache, try to hit it.
		if cache.Cached {
			// use cache! manually place cache contents into destination
			datS, datE := withSize(dst, uint64(len(cache.Serialized)))
			copy((*dst)[datS:datE], cache.Serialized)
		} else {
			// The start of the data will be the end of the current destination data
			datS := uint32(len(*dst))
			// serialize like normal
			sszSerialize(v, dst)
			// fill cache using data that was placed in the destination
			cache.Serialized = (*dst)[datS:]
			cache.Cached = true
		}
	} else {
		// no cache available, handle like normal
		sszSerialize(v, dst)
	}
}


type offsetVariableRef struct {
	offsetIndex uint32
	fieldIndex int
}

func sszSerialize(v reflect.Value, dst *[]byte) {
	switch v.Kind() {
	case reflect.Ptr:
		sszSerialize(v.Elem(), dst)
	case reflect.Uint8: // "uintN"
		s, _ := withSize(dst, 1)
		(*dst)[s] = byte(v.Uint())
	case reflect.Uint32: // "uintN"
		s, e := withSize(dst, 4)
		binary.LittleEndian.PutUint32((*dst)[s:e], uint32(v.Uint()))
	case reflect.Uint64: // "uintN"
		s, e := withSize(dst, 8)
		binary.LittleEndian.PutUint64((*dst)[s:e], uint64(v.Uint()))
	case reflect.Bool: // "bool"
		s, _ := withSize(dst, 1)
		if v.Bool() {
			(*dst)[s] = 1
		} else {
			(*dst)[s] = 0
		}
	case reflect.Array, reflect.Slice: // "vector", "list"
		if elemT := v.Type().Elem(); elemT.Kind() == reflect.Uint8 {
			// just directly copy over the bytes if we can
			s, e := withSize(dst, uint64(v.Len()))
			if v.Kind() == reflect.Array {
				v = v.Slice(0, v.Len())
			}
			copy((*dst)[s:e], v.Bytes())
		} else if isFixedSize(elemT) {
			// not a bytes array, but fixed size still.
			for i, size := 0, v.Len(); i < size; i++ {
				sszSerializeMaybeCached(v.Index(i), dst)
			}
		} else {
			// fixed part: allocate offsets
			absStart := len(*dst)
			s, _ := withSize(dst, BYTES_PER_LENGTH_OFFSET * uint64(v.Len()))
			// write the variable parts, and back-fill the offsets
			for i, size := 0, v.Len(); i < size; i++ {
				// offset points to end of previous data
				// offset is located at start of fixed + (BYTES_PER_LENGTH_OFFSET * i)
				binary.LittleEndian.PutUint32((*dst)[s:s+BYTES_PER_LENGTH_OFFSET], uint32(len(*dst) - absStart))
				s += BYTES_PER_LENGTH_OFFSET
				// append to dst
				sszSerializeMaybeCached(v.Index(i), dst)
			}
		}
	case reflect.Struct: // "container"
		vType := v.Type()
		varParts := make([]offsetVariableRef, 0)
		absStart := len(*dst)
		for i, size := 0, v.NumField(); i < size; i++ {
			field := vType.Field(i)
			if tags.HasFlag(&field, SSZ_TAG, OMIT_FLAG) {
				continue
			}
			if isFixedSize(field.Type) {
				// append to dst
				sszSerializeMaybeCached(v.Field(i), dst)
			} else {
				// just create space for the offset (to be written later)
				s, _ := withSize(dst, BYTES_PER_LENGTH_OFFSET)
				varParts = append(varParts, offsetVariableRef{uint32(s), i})
			}
		}
		for _, fp := range varParts {
			// offset points to end of previous data
			binary.LittleEndian.PutUint32((*dst)[fp.offsetIndex:fp.offsetIndex+BYTES_PER_LENGTH_OFFSET], uint32(len(*dst) - absStart))
			// append to dst
			sszSerializeMaybeCached(v.Field(fp.fieldIndex), dst)
		}
	default:
		panic("ssz encoding: unsupported value kind: " + v.Kind().String())
	}
}

// constructs a merkle_root of the given data, but truncates last element (i.e. ignored, not part of the root)
func SigningRoot(input interface{}) Root {
	return signingRoot(reflect.ValueOf(input))
}

func signingRoot(v reflect.Value) Root {
	switch v.Kind() {
	case reflect.Ptr:
		return signingRoot(v.Elem())
	case reflect.Struct:
		data := composeStructRootData(v)
		if len(data) <= 1 {
			panic("taking signing-root of single/non-field struct")
		}
		return merkle.MerkleRoot(data[:len(data)-1])
	default:
		panic("input of signing root is not a struct")
	}
}

// constructs a merkle_root of the given data
func HashTreeRoot(input interface{}) Root {
	return sszHashTreeRoot(reflect.ValueOf(input), nil)
}

func isFixedSize(vt reflect.Type) bool {
	switch vt.Kind() {
	case reflect.Ptr:
		return isFixedSize(vt.Elem())
	case reflect.Uint8, reflect.Uint32, reflect.Uint64, reflect.Bool:
		return true
	case reflect.Slice:
		return false
	case reflect.Array:
		return isFixedSize(vt.Elem())
	case reflect.Struct:
		// We want to encode the struct in the most minimal way,
		//  if it can be fixed length, make it fixed length.
		for i, length := 0, vt.NumField(); i < length; i++ {
			field := vt.Field(i)
			if tags.HasFlag(&field, SSZ_TAG, OMIT_FLAG) {
				continue
			}
			if !isFixedSize(vt.Field(i).Type) {
				return false
			}
		}
		return true
	default:
		panic("is-fixed size: unsupported value type: " + vt.String())
	}
}

func isBasicType(vt reflect.Type) bool {
	switch vt.Kind() {
	case reflect.Uint8, reflect.Uint32, reflect.Uint16, reflect.Uint64: // No uint128 and uint256 used
		return true
	case reflect.Bool:
		return true
	default:
		return false
	}
}

func basicListRoot(v reflect.Value, compoundCache *SSZCompoundCache) Root {
	if compoundCache != nil {
		serializeFn := func(dst []byte, index uint64) {
			sszSerialize(v.Index(int(index)), &dst)
		}
		// update and use cache for merkleization
		return compoundCache.UpdateAndMerkleize(serializeFn)
	} else {
		return merkle.MerkleRoot(sszPack(v))
	}
}

// only call this for slices and arrays, not structs
func nonBasicListRoot(v reflect.Value, compoundCache *SSZCompoundCache) Root {
	if compoundCache != nil {
		serializeFn := func(dst []byte, index uint64) {
			hash := sszHashTreeRoot(v.Index(int(index)), nil)
			copy(dst, hash[:])
		}
		// update and use cache for merkleization
		return compoundCache.UpdateAndMerkleize(serializeFn)
	} else {
		items := v.Len()
		data := make([]Root, items)
		for i := 0; i < items; i++ {
			data[i] = sszHashTreeRoot(v.Index(i), nil)
		}
		return merkle.MerkleRoot(data)
	}
}

func composeStructRootData(v reflect.Value) []Root {
	fields := v.NumField()
	data := make([]Root, 0, fields)
	vType := v.Type()
	for i := 0; i < fields; i++ {
		field := vType.Field(i)
		if !tags.HasFlag(&field, SSZ_TAG, OMIT_FLAG) {
			cacheFieldName := field.Name + "SSZCache"
			_, ok := vType.FieldByName(cacheFieldName)
			fieldV := v.Field(i)
			if ok {
				cacheV := v.FieldByName(cacheFieldName)
				elemCache := cacheV.Interface().(SSZCompoundCache)
				// cache may be nil
				data = append(data, sszHashTreeRoot(fieldV, &elemCache))
			} else {
				data = append(data, sszHashTreeRoot(fieldV, nil))
			}
		}
	}
	return data
}

func structRoot(v reflect.Value) Root {
	return merkle.MerkleRoot(composeStructRootData(v))
}

// Compute hash tree root for a value
func sszHashTreeRoot(v reflect.Value, compoundCache *SSZCompoundCache) Root {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return Root{}
		}
		return sszHashTreeRoot(v.Elem(), compoundCache)
	// "basic object? -> pack and merkle_root
	case reflect.Uint8, reflect.Uint32, reflect.Uint64, reflect.Bool:
		return merkle.MerkleRoot(sszPack(v))
	// "array of basic items? -> pack and merkle_root
	// "array of non basic items? -> take the merkle root of every element
	// 		(if it aligns in a single chunk, it will just be as-is, not hashed again, see merklezeition)
	case reflect.Array:
		if isBasicType(v.Type().Elem()) {
			return basicListRoot(v, compoundCache)
		} else {
			return nonBasicListRoot(v, compoundCache)
		}
	case reflect.Slice:
		if isBasicType(v.Type().Elem()) {
			return sszMixInLength(basicListRoot(v, compoundCache), uint64(v.Len()))
		} else {
			return sszMixInLength(nonBasicListRoot(v, compoundCache), uint64(v.Len()))
		}
	case reflect.Struct:
		return structRoot(v)
	default:
		panic("tree-hash: unsupported value kind: " + v.Kind().String())
	}
}

func sszPack(input reflect.Value) []Root {
	var serialized []byte
	if cache := getSerializeCache(input); cache != nil && cache.Cached {
		// use cache!
		serialized = cache.Serialized
	} else {
		serialized = make([]byte, 0)

		switch input.Kind() {
		case reflect.Array, reflect.Slice:
			for i, length := 0, input.Len(); i < length; i++ {
				sszSerialize(input.Index(i), &serialized)
			}
		case reflect.Struct:
			vType := input.Type()
			for i, length := 0, input.NumField(); i < length; i++ {
				field := vType.Field(i)
				if !tags.HasFlag(&field, SSZ_TAG, OMIT_FLAG) {
					sszSerialize(input.Field(i), &serialized)
				}
			}
		default:
			sszSerialize(input, &serialized)
		}

		// check if there was a cache, but it was just empty/invalid. In that case, fill it
		if cache != nil {
			cache.Serialized = serialized
			cache.Cached = true
		}
	}

	// floored: handle all normal chunks first
	flooredChunkCount := len(serialized) / 32
	// ceiled: include any partial chunk at end as full chunk (with padding)
	out := make([]Root, (len(serialized)+31)/32)
	for i := 0; i < flooredChunkCount; i++ {
		copy(out[i][:], serialized[i<<5:(i+1)<<5])
	}
	// if there is a partial chunk at the end, handle it as a special case:
	if len(serialized)&31 != 0 {
		copy(out[flooredChunkCount][:len(serialized)&0x1F], serialized[flooredChunkCount<<5:])
	}
	return out
}

func sszMixInLength(data Root, length uint64) Root {
	lengthInput := Root{}
	binary.LittleEndian.PutUint64(lengthInput[:], length)
	return merkle.MerkleRoot([]Root{data, lengthInput})
}
