package spec_testing

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

type TestCase interface {
	Run(t *testing.T)
}

type CaseAllocator func() interface{}

type TestCasesHolder struct {
	CaseAlloc CaseAllocator
	Cases []TestCase
}

func (holder *TestCasesHolder) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var rawCases []interface{}
	// read raw YAML into parsed but untyped structure
	if err := unmarshal(&rawCases); err != nil {
		return err
	}

	// python style data -> go style data
	transformed := decodeList(rawCases)

	holder.Cases = make([]TestCase, 0, len(transformed))

	for _, transformedCase := range transformed {
		// Hack: encode, and decode. Bit slow, but makes it easy to load the data into a fully typed struct
		encodedCase, err := json.Marshal(transformedCase)
		if err != nil {
			return errors.New(fmt.Sprintf("cannot parse spec data: %v", err))
		}
		caseTyped := holder.CaseAlloc()
		if err := json.Unmarshal(encodedCase, caseTyped); err != nil {
			return errors.New(fmt.Sprintf("cannot decode test-case: %v", err))
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
	if err := yaml.Unmarshal(yamlBytes, suite); err != nil {
		return nil, err
	}
	return suite, nil
}

func (suite *TestSuite) Run(t *testing.T) {
	testRunTitle := fmt.Sprintf("%s > %s ~ %s [%d cases]",
		suite.Runner, suite.Handler, suite.Title, len(suite.TestCases.Cases))
	t.Run(testRunTitle, func(t *testing.T) {
		for i, testCase := range suite.TestCases.Cases {
			t.Run(testRunTitle + fmt.Sprintf(" case #%d", i), func(t *testing.T) {
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
		t.Error(err)
	}

	for _, path := range suitePaths {
		suite, err := LoadSuite(path, caseAlloc)
		if err != nil {
			t.Error(err)
		}
		suite.Run(t)
	}
}