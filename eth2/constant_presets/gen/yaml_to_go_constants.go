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

type PresetEntry struct {
	Key string
	Value string
}

type ContstantsPreset struct {
	Name string
	Entries []PresetEntry
}

func main() {
	var presetsDirPath string
	flag.StringVar(&presetsDirPath, "path", "", "The file path to the directory containing yaml constant presets")
	flag.Parse()

	templ := template.Must(template.New("constants_file").Parse(constantsFileTemplate))

	if err := filepath.Walk(presetsDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		extension := filepath.Ext(path)
		if extension != ".yaml" {
			return filepath.SkipDir
		}

		presetName := filepath.Base(path)
		presetName = presetName[:len(presetName)-len(".yaml")]

		yamlBytes, err := ioutil.ReadFile(path)
		rawPreset := make(map[string]interface{})
		if err := yaml.Unmarshal(yamlBytes, yamlBytes); err != nil {
			return err
		}

		preset := ContstantsPreset{
			Name: presetName,
			Entries: make([]PresetEntry, 0, len(rawPreset)),
		}
		for k, v := range rawPreset {
			formattedValue := ""
			if strV, ok := v.(string); ok {
				if intV, err := strconv.ParseInt(strV, 0, 64); err != nil {
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

			preset.Entries = append(preset.Entries, PresetEntry{
				Key: k,
				Value: formattedValue,
			})
		}

		// TODO
		f, err := os.Create(filepath.Join("todo", presetName + ".go"))
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

var constantsFileTemplate = `
// +build preset_{{.Name}}

package constant_presets

// TODO iterate entries

const {{.Key}} = {{.Value}}

`
