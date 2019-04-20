# SSZ

Implementation of the [SSZ spec for ETH 2.0](https://github.com/ethereum/eth2.0-specs/blob/dev/specs/simple-serialize.md)

## Caching

### Serialization caching

You can embed a `SerializationCacher` in your struct, and the ssz functions will automatically fill and use the cache.
To reset the cache, set `myStruct.Cached = false` (field is from embedded cacher struct).

### Compound caching

Lists do not provide a straigt-forward way to add a cache for merkleization, 
 you cannot embed functionality like with structs. Because of this, we have to do a small work-around:
 a separate `SSZCompoundCache`.
This should be located in the struct containing the list field that needs to be cached.

To reset a part of the cache matching an item in the list, call `myCache.SetChanged(<index>)`.

TODO: this is a work in progress, integration pattern (Embed vs interface vs reflection) unclear.

## Omit fields

You can choose to ignore any struct field during serialization and hash-tree-root by tagging it with `ssz:"omit"`

## Ignore last field ("signed root")

`SignedRoot()` computes the hash-tree-root, ignoring the last field of the struct.

## TODO

- make compound cache work for structs as well. (less useful however, structs are not as big as lists)
- make embeddable hash-tree-root cache for caching roots
