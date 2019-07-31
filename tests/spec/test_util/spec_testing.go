package test_util

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/core"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type ConfigMismatchError struct {
	Config string
}

func (confErr ConfigMismatchError) Error() string {
	return fmt.Sprintf("cannot load suite for config: %s, current config is: %s", confErr.Config, core.PRESET_NAME)
}

type TestCase interface {
	Run(t *testing.T)
}

type NamedTestCase struct {
	TestCase
	Name string
}

type TestPartReader func(name string) (r io.Reader, size uint64)

// Runs a test case
type CaseRunner func(t *testing.T, readPart TestPartReader)


func Check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func RunHandler(t *testing.T, handlerPath string, caseRunner CaseRunner, config string) {
	// general config is allowed
	if config != core.PRESET_NAME && config != "general" {
		t.Logf("Config %s does not match current config %s, " +
			"skipping handler %s", config, core.PRESET_NAME, handlerPath)
	}

	// get the current path, go to the root, and get the tests path
	_, filename, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(filepath.Dir(filename))
	handlerAbsPath := filepath.Join(basepath, "eth2.0-spec-tests", "tests", config, filepath.FromSlash(handlerPath))

	forEachDir := func(t *testing.T, path string, callItem func(t *testing.T, path string)) {
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
		partReader := func(name string) (r io.Reader, size uint64) {
			partPath := filepath.Join(path, name)
			partInfo, err := os.Stat(partPath)
			Check(t, err)
			// get the size
			size = uint64(partInfo.Size())
			f, err := os.Open(partPath)
			Check(t, err)
			r = f
			return
		}
		caseRunner(t, partReader)
	}

	runSuite := func (t *testing.T, path string) {
		forEachDir(t, path, runTest)
	}

	forEachDir(t, handlerAbsPath, runSuite)
}
