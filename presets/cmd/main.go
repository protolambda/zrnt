package main

import (
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type ContstantsPreset struct {
	Name string
	Entries []string
}

func main() {
	var presetsDirPath, outputDirPath string
	flag.StringVar(&presetsDirPath, "presets-dir", "", "The file path to the directory containing yaml constant presets")
	flag.StringVar(&outputDirPath, "output-dir", "", "The file path to the directory to output generated Go code to")
	flag.Parse()

	templ := template.Must(template.New("constants_file").Parse(constantsFileTemplate))

	if err := filepath.Walk(presetsDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("processing preset file", path)

		extension := filepath.Ext(path)
		if extension != ".yaml" {
			return nil
		}

		presetName := filepath.Base(path)
		presetName = presetName[:len(presetName)-len(".yaml")]

		fmt.Println("processing preset", presetName)

		yamlBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		rawPreset := make(map[string]interface{})
		if err := yaml.Unmarshal(yamlBytes, &rawPreset); err != nil {
			return err
		}

		preset := ContstantsPreset{
			Name: presetName,
			Entries: make([]string, 0, len(rawPreset)),
		}
		for k, v := range rawPreset {
			formattedValue := ""
			formattedStart := "const " + k + " = "
			if strV, ok := v.(string); ok {
				if intV, err := strconv.ParseInt(strV, 0, 64); err == nil {
					formattedValue = fmt.Sprintf("%d", intV)
				} else if strings.HasPrefix(strV, "0x") {
					strNibbles := strV[2:]
					if len(strNibbles) % 2 != 0 {
						return errors.New(fmt.Sprintf("invalid constant in %s, %s has value %s", presetName, k, strV))
					}
					byteCount := len(strNibbles)
					formattedValue = fmt.Sprintf("[%d]byte{", byteCount)
					for i := 0; i < len(strNibbles); i += 2 {
						formattedValue += "0x" + strNibbles[i:i+2]
						if i + 2 < len(strNibbles) {
							formattedValue += ", "
						}
					}
					formattedValue += "}"
					// arrays cannot be constants
					formattedStart = "var " + k + " = "
				} else {
					return errors.New(fmt.Sprintf("could not convert string formatted value in %s, key: %s, value: %s", presetName, k, strV))
				}
			} else {
				if uintV, ok := v.(uint64); ok {
					formattedValue = fmt.Sprintf("%d", uintV)
				} else if intV, ok := v.(int); ok {
					formattedValue = fmt.Sprintf("%d", intV)
				} else {
					return errors.New(fmt.Sprintf("could not convert non-string formatted value in %s, key: %s %T", presetName, k, v))
				}
			}

			preset.Entries = append(preset.Entries, formattedStart + formattedValue)
		}

		outPath := filepath.Join(outputDirPath, presetName + ".go")
		fmt.Printf("writing constants preset %s to %s\n", presetName, outPath)
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if err := templ.Execute(f, preset); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

}

var constantsFileTemplate = `// +build preset_{{.Name}}

package constant_presets

const PRESET_NAME string = "{{.Name}}"

{{ range .Entries }}
{{.}}
{{ end }}
`
