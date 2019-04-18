package spec_testing

import (
	"errors"
	"fmt"
	. "github.com/mitchellh/mapstructure"
	"github.com/protolambda/zrnt/constant_presets"
	"github.com/protolambda/zrnt/eth2/util/hex"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type TestSuite struct {
	// Display name for the test suite
	Title string `yaml:"title"`
	// Summarizes the test suite
	Summary string `yaml:"summary"`

	// TODO: forks is ignored for now, no special behaviors yet.

	// References applicable forks.
	// Runner decides what to do: Run for each fork,
	// or run for all at once, each fork transition, etc.
	Forks []string `yaml:"forks"`
	// Reference to a fork definition file, without extension. Used to determine the forking timeline
	ForksTimeline string `yaml:"forks_timeline"`
	// Used to determine which set of constants to run (possibly compile time) with
	Config  string `yaml:"config"`
	Runner  string `yaml:"runner"`
	Handler string `yaml:"handler"`

	TestCases TestCasesHolder `yaml:"test_cases"`
}

type ConfigMismatchError struct {
	Config string
}

func (confErr ConfigMismatchError) Error() string {
	return fmt.Sprintf("cannot load suite for config: %s, current config is: %s", confErr.Config, constant_presets.CONFIG)
}

type TestSuiteLoader struct {
	Config  string `yaml:"config"`
}


type TestCase interface {
	Run(t *testing.T)
}

type CaseAllocator func(raw interface{}) interface{}

type TestCasesHolder struct {
	CaseAlloc CaseAllocator
	Cases []TestCase
}

func decodeHook(s reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t.Kind() == reflect.Slice && t.Elem().Kind() != reflect.Uint8 {
		return data, nil
	}
	if s.Kind() != reflect.String {
		return data, nil
	}
	strData := data.(string)
	if t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Uint8 {
		res := reflect.New(t).Elem()
		sliceRes := res.Slice(0, t.Len()).Interface()
		err := hex.DecodeHex([]byte(strData), sliceRes.([]byte))
		return res.Interface(), err
	}
	if t.Kind() == reflect.Uint64 {
		return strconv.ParseUint(strData, 10, 64)
	}
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		inBytes := []byte(strData)
		_, byteCount := hex.DecodeHexOffsetAndLen(inBytes)
		res := make([]byte, byteCount, byteCount)
		err := hex.DecodeHex([]byte(strData), res)
		return res, err
	}
	return data, nil
}

func (holder *TestCasesHolder) UnmarshalYAML(unmarshal func(interface{}) error) error {
	rawCases := make([]interface{}, 0)
	// read raw YAML into parsed but untyped structure
	if err := unmarshal(&rawCases); err != nil {
		return err
	}

	// python style data -> go style data
	transformed := decodeList(rawCases)

	holder.Cases = make([]TestCase, 0, len(transformed))

	for _, transformedCase := range transformed {
		caseTyped := holder.CaseAlloc(transformedCase)

		var md Metadata
		config := &DecoderConfig{
			DecodeHook: decodeHook,
			Metadata: &md,
			WeaklyTypedInput: false,
			Result:           caseTyped,
		}

		decoder, err := NewDecoder(config)
		if err != nil {
			return err
		}

		if err := decoder.Decode(transformedCase); err != nil {
			return errors.New(fmt.Sprintf("cannot decode test-case: %v", err))
		}
		if len(md.Unused) > 0 {
			return errors.New(fmt.Sprintf("unused keys: %s", strings.Join(md.Unused, ", ")))
		}
		asTestCase, ok := caseTyped.(TestCase)
		if !ok {
			return errors.New(fmt.Sprintf("cannot hold parsed test as generic test case: %T", caseTyped))
		}
		holder.Cases = append(holder.Cases, asTestCase)
	}
	return nil
}

func LoadSuite(path string, caseAlloc CaseAllocator) (*TestSuite, error) {
	suite := new(TestSuite)
	suite.TestCases.CaseAlloc = caseAlloc
	yamlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read test cases file %s %v", path, err))
	}
	suiteLoader := new(TestSuiteLoader)
	if err := yaml.Unmarshal(yamlBytes, suiteLoader); err != nil {
		return nil, err
	}
	// skip test-suites with different configurations
	if suiteLoader.Config != constant_presets.CONFIG {
		return nil, ConfigMismatchError{suiteLoader.Config}
	}
	if err := yaml.Unmarshal(yamlBytes, suite); err != nil {
		return nil, err
	}
	return suite, nil
}

func (suite *TestSuite) Run(t *testing.T) {
	t.Run(suite.Title, func(t *testing.T) {
		for i, testCase := range suite.TestCases.Cases {
			t.Run(fmt.Sprintf("case #%d", i), func(t *testing.T) {
				testCase.Run(t)
			})
		}
	})
}

func RunSuitesInPath(path string, caseAlloc CaseAllocator, t *testing.T) {
	suitePaths := make([]string, 0)

	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("processing file", path)

		extension := filepath.Ext(path)
		if extension != ".yaml" {
			// dig deeper (i.e. do not return directory-skip error)
			return nil
		}

		suitePaths = append(suitePaths, path)

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	for _, path := range suitePaths {
		suite, err := LoadSuite(path, caseAlloc)
		if confErr, ok := err.(ConfigMismatchError); ok {
			t.Logf("Config %s of test-suite does not match current config %s, skipping suite %s", confErr.Config, constant_presets.CONFIG, path)
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		suite.Run(t)
	}
}
