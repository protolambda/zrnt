# constant presets: generated code

To configure constants during compile time, the configurations are transformed in Go files using the generation functionality of Go.
 Each of the different configurations gets it's own build constraint, prefixed with `preset_`: e.g. `preset_mainnet`

The `constants.go` in the core baecon package makes these constants available to the rest of the code, and types them.
