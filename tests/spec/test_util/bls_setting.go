package test_util

import (
	"gopkg.in/yaml.v2"
	"testing"
)

const (
	BlsOptional = 0
	BlsRequired = 1
	BlsIgnored  = 2
)

type BLSMeta struct {
	BlsSetting  int `yaml:"bls_setting"`
}

func HandleBLS(testRunner CaseRunner) CaseRunner {
	return func(t *testing.T, readPart TestPartReader) {
		part := readPart("meta.yaml")
		if part.Exists() {
			meta := BLSMeta{}
			dec := yaml.NewDecoder(part)
			Check(t, dec.Decode(&meta))
			Check(t, part.Close())
			// TODO: change to environment check once BLS is supported by ZRNT
			if meta.BlsSetting == BlsRequired {
				t.Skip("skipping BLS-only test")
				return
			}
		}
		testRunner(t, readPart)
	}
}
