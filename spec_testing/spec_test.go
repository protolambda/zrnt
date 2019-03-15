package spec_testing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
)

type SpecTests struct {
	Title string `yaml:"title"`
	Summary string `yaml:"summary"`
	TestSuite string `yaml:"test_suite"`
	Fork string `yaml:"fork"`
	Version string `yaml:"version"`
	TestCases []map[interface{}]interface{} `yaml:"test_cases"`
}

func TestAllCases(t *testing.T)  {

	configPath := "test_file.yaml"

	yamlBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("cannot read test cases file %s %v", configPath, err)
	}

	suite := SpecTests{}

	if err := yaml.Unmarshal(yamlBytes, &suite); err != nil {
		t.Fatalf("cannot read spec test case yaml: %v", err)
	}

	for _, caseData := range suite.TestCases {
		configData := caseData["config"].(map[interface{}]interface{})
		config := make([]string, 0, len(configData))
		for k, v := range configData {
			config = append(config, fmt.Sprintf("-X github.com/protolambda/zrnt/eth2/beacon.%s=%d", k, v))
		}
		delete(caseData, "config")
		decoded := DecodeSpecFormat(caseData)
		t.Logf("%v", decoded)
		encodedCase, err := json.Marshal(decoded)
		if err != nil {
			t.Error(err)
			continue
		}
		b64encoded := base64.StdEncoding.EncodeToString(encodedCase)
		t.Log(string(b64encoded))
		cmd := exec.Command("go", "test", "-run", "^TestSpecCase", "-ldflags", "\""+strings.Join(config, " ")+"\"", "-args", b64encoded)
		fmt.Println(strings.Join(cmd.Args, " "))
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Error(fmt.Sprint(err) + ": " + stderr.String() + out.String())
		}
	}
}