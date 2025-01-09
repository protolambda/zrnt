package test_util

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang/snappy"
	"github.com/protolambda/ztyp/codec"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type TestPart interface {
	io.Reader
	io.Closer
	Exists() bool
}

type TestPartReader interface {
	Part(name string) TestPart
	Spec() *common.Spec
}

// Runs a test case
type CaseRunner func(t *testing.T, forkName ForkName, readPart TestPartReader)

func Check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type testPartFile struct {
	*os.File
}

func (p *testPartFile) Size() (uint64, error) {
	partInfo, err := p.Stat()
	if err != nil {
		return 0, err
	} else {
		return uint64(partInfo.Size()), nil
	}
}

func (p *testPartFile) Exists() bool {
	return p.File != nil
}

type partAndSpec struct {
	readPart func(name string) TestPart
	spec     *common.Spec
}

func (s *partAndSpec) Part(name string) TestPart {
	return s.readPart(name)
}

func (s *partAndSpec) Spec() *common.Spec {
	return s.spec
}

func RunHandler(t *testing.T, handlerPath string, caseRunner CaseRunner, spec *common.Spec, fork ForkName) {
	// get the current path, go to the root, and get the tests path
	_, filename, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(filepath.Dir(filename))
	handlerAbsPath := filepath.Join(basepath, "eth2.0-spec-tests", "tests",
		spec.PRESET_BASE, string(fork), filepath.FromSlash(handlerPath))

	forEachDir := func(t *testing.T, path string, callItem func(t *testing.T, path string)) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Skipf("missing tests: %s", path)
		} else {
			Check(t, err)
		}
		items, err := ioutil.ReadDir(path)
		Check(t, err)
		for _, item := range items {
			if item.IsDir() {
				t.Run(item.Name(), func(t *testing.T) {
					callItem(t, filepath.Join(path, item.Name()))
				})
			}
		}
	}

	runTest := func(t *testing.T, path string) {
		//t.Parallel()
		partReader := func(name string) TestPart {
			partPath := filepath.Join(path, name)
			if _, err := os.Stat(partPath); os.IsNotExist(err) {
				return &testPartFile{File: nil}
			} else {
				f, err := os.Open(partPath)
				Check(t, err)
				return &testPartFile{File: f}
			}
		}
		caseRunner(t, fork, &partAndSpec{readPart: partReader, spec: spec})
	}

	runSuite := func(t *testing.T, path string) {
		//t.Parallel()
		forEachDir(t, path, runTest)
	}

	t.Run(handlerPath, func(t *testing.T) {
		//t.Parallel()
		forEachDir(t, handlerAbsPath, runSuite)
	})
}

func LoadSpecObj(t *testing.T, name string, dst common.SpecObj, readPart TestPartReader) bool {
	t.Helper()
	p := readPart.Part(name + ".ssz_snappy")
	if p.Exists() {
		data, err := ioutil.ReadAll(p)
		Check(t, err)
		Check(t, p.Close())
		uncompressed, err := snappy.Decode(nil, data)
		Check(t, err)
		spec := readPart.Spec()
		Check(t, dst.Deserialize(spec, codec.NewDecodingReader(bytes.NewReader(uncompressed), uint64(len(uncompressed)))))
		return true
	} else {
		return false
	}
}

func LoadSSZ(t *testing.T, name string, dst codec.Deserializable, readPart TestPartReader) bool {
	t.Helper()
	p := readPart.Part(name + ".ssz_snappy")
	if p.Exists() {
		data, err := ioutil.ReadAll(p)
		Check(t, err)
		Check(t, p.Close())
		uncompressed, err := snappy.Decode(nil, data)
		Check(t, err)
		Check(t, dst.Deserialize(codec.NewDecodingReader(bytes.NewReader(uncompressed), uint64(len(uncompressed)))))
		return true
	} else {
		return false
	}
}
