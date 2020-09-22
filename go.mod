module github.com/protolambda/zrnt

go 1.14

require (
	github.com/herumi/bls-eth-go-binary v0.0.0-20200522010937-01d282b5380b
	github.com/minio/sha256-simd v0.1.0
	github.com/protolambda/messagediff v1.3.0
	github.com/protolambda/zssz v0.1.5
	github.com/protolambda/ztyp v0.0.2
	gopkg.in/yaml.v2 v2.2.2
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace github.com/protolambda/ztyp => ../ztyp
replace github.com/protolambda/messagediff => ../messagediff
