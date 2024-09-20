module github.com/protolambda/zrnt

go 1.21

require (
	github.com/golang/snappy v0.0.3
	github.com/kilic/bls12-381 v0.1.0
	github.com/minio/sha256-simd v0.1.0
	github.com/protolambda/bls12-381-util v0.1.0
	github.com/protolambda/messagediff v1.4.0
	github.com/protolambda/ztyp v0.2.2
	gopkg.in/yaml.v3 v3.0.0
)

require (
	github.com/holiman/uint256 v1.2.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

replace (
	github.com/ethereum/go-ethereum => github.com/overprotocol/kairos v0.0.0-20240604101705-10a725a19997
	github.com/protolambda/zrnt => github.com/superblock-dev/zrnt v0.0.0-20240208044402-a0e6af5cc000
)
