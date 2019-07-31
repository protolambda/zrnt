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
		r, _ := readPart("meta.yaml")
		meta := BLSMeta{}
		dec := yaml.NewDecoder(r)
		Check(t, dec.Decode(meta))
		// TODO: change to environment check once BLS is supported by ZRNT
		if meta.BlsSetting == BlsRequired {
			t.Skip("skipping BLS-only test")
		}
		testRunner(t, readPart)
	}
}
