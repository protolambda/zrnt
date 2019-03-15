package spec_testing

type SpecTestsSuite struct {
	Title string `yaml:"title"`
	Summary string `yaml:"summary"`
	TestSuite string `yaml:"test_suite"`
	Fork string `yaml:"fork"`
	Version string `yaml:"version"`
	TestCases []map[interface{}]interface{} `yaml:"test_cases"`
}
