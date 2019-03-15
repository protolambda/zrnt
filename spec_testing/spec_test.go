package spec_testing

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
)


func TestAllCases(t *testing.T)  {

	suiteFilePath := "test_file.yaml"

	yamlBytes, err := ioutil.ReadFile(suiteFilePath)
	if err != nil {
		t.Fatalf("cannot read suite file %s %v", suiteFilePath, err)
	}

	suite := SpecTestsSuite{}

	if err := yaml.Unmarshal(yamlBytes, &suite); err != nil {
		t.Fatalf("cannot read spec test case yaml: %v", err)
	}

	for i, caseData := range suite.TestCases {
		configData := caseData["config"].(map[interface{}]interface{})
		config := make([]string, 0, len(configData))
		for k, v := range configData {
			config = append(config, fmt.Sprintf("-X github.com/protolambda/zrnt/eth2/beacon.%s=%d", k, v))
		}
		cmd := exec.Command("go", "test", "-run", "^TestSpecCase", "-ldflags=\""+strings.Join(config, " ")+"\"", "-args", suiteFilePath, fmt.Sprintf("%d", i))
		fmt.Println(strings.Join(cmd.Args, " "))
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Errorf("failed to run test %s %d.\n%v\n%v", suiteFilePath, i, stderr.String(), out.String())
		}
	}
}