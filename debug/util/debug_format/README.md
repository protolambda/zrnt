# Debug Format

This package is for a special serialization format used to debug with:
 it adds a `_hash_tree_root` field for every original field, with the SSZ hash tree root 
 (encoded as 64 char hex string, no `0x` prefix).

The format is easy to wrap in other formats: e.g. the JSON encoder is just 1 line.

Encode raw (returns a structure of slices/maps, to encode with another encoder:

```golang
encoded := debug_format.Encode(myValue)
```

To use JSON directly:

```golang
json_bytes, err := debug_format.MarshalJSON(myValue)
```
