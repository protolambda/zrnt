module github.com/superblock-dev/zrnt

go 1.16

require (
	github.com/golang/snappy v0.0.3
	github.com/kilic/bls12-381 v0.1.0
	github.com/minio/sha256-simd v0.1.0
	github.com/protolambda/bls12-381-util v0.0.0-20210720105258-a772f2aac13e
	github.com/protolambda/messagediff v1.4.0
	github.com/protolambda/zrnt v0.30.0
	github.com/protolambda/ztyp v0.2.2
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace (
	github.com/ethereum/go-ethereum => github.com/superblock-dev/kairos v0.0.0-20230725040008-76a94c12bf9b
	github.com/protolambda/zrnt => github.com/superblock-dev/zrnt v0.0.0-20240208044402-a0e6af5cc000
)