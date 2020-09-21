package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type ContstantsPresetData struct {
	Name      string
	NeedsUtil bool
	Entries   []string
}

func hexStrToLiteralStr(hex string, typeName string, expByteLen int) (string, error) {
	if strings.HasPrefix(hex, "0x") {
		strNibbles := hex[2:]
		if len(strNibbles)%2 != 0 {
			return "", fmt.Errorf("invalid value: %s", hex)
		}
		hex = strNibbles
	}
	byteCount := len(hex) / 2
	if expByteLen >= 0 {
		if expByteLen != byteCount {
			return "", fmt.Errorf("unexpected byte length %d, expected %d", byteCount, expByteLen)
		}
	}
	if typeName == "" {
		typeName = fmt.Sprintf("[%d]byte", byteCount)
	}
	formattedValue := typeName + "{"
	for i := 0; i < len(hex); i += 2 {
		formattedValue += "0x" + hex[i:i+2]
		if i+2 < len(hex) {
			formattedValue += ", "
		}
	}
	formattedValue += "}"
	return formattedValue, nil
}

func buildPreset(path string) (*ContstantsPresetData, error) {
	presetName := filepath.Base(path)
	presetName = presetName[:len(presetName)-len(filepath.Ext(path))]

	fmt.Println("processing preset", presetName)

	yamlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rawPreset yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &rawPreset); err != nil {
		return nil, err
	}

	preset := ContstantsPresetData{
		Name:      presetName,
		NeedsUtil: false,
		Entries:   make([]string, 0),
	}
	if len(rawPreset.Content) != 1 || rawPreset.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("bad config root content")
	}
	items := rawPreset.Content[0].Content
	for i := 0; i < len(items); i += 2 {
		rk := items[i]
		if rk.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("key %s (index %d) is invalid kind: %d", rk.Value, i/2, rk.Kind)
		}
		k := rk.Value

		formattedValue := ""
		formattedStart := k + ": "

		rv := items[i+1]
		switch rv.Kind {
		case yaml.ScalarNode:
			v := rv.Value
			if strings.HasPrefix(k, "DOMAIN_") {
				formattedValue, err = hexStrToLiteralStr(v, "BLSDomainType", 4)
				if err != nil {
					return nil, fmt.Errorf("bad domain value: %s: %s", k, v)
				}
			} else if strings.HasSuffix(k, "_FORK_VERSION") {
				formattedValue, err = hexStrToLiteralStr(v, "Version", 4)
				if err != nil {
					return nil, fmt.Errorf("bad version value: %s: %s", k, v)
				}
			} else if k == "BLS_WITHDRAWAL_PREFIX" {
				formattedValue, err = hexStrToLiteralStr(v, "Version", 1)
				if err != nil {
					return nil, fmt.Errorf("bad withdrawal prefix value: %s: %s", k, v)
				}
			} else if k == "DEPOSIT_CONTRACT_ADDRESS" {
				formattedValue, err = hexStrToLiteralStr(v, "[20]byte", 20)
				if err != nil {
					return nil, fmt.Errorf("bad contract address value: %s: %s", k, v)
				}
			} else if k == "CUSTODY_PRIME" {
				var x big.Int
				if err := x.UnmarshalText([]byte(v)); err != nil {
					return nil, fmt.Errorf("invalid big int: %s: %s, error: %v", k, v, err)
				}
				preset.NeedsUtil = true
				formattedValue = fmt.Sprintf("util.MustBigInt(\"%s\")", x.String())
			} else {
				// format values nicely: if starts with 0x, keep it.
				// Or if it's a high number (Configs generally do decimal representation more often)
				if intV, err := strconv.ParseUint(v, 0, 64); err == nil {
					if strings.HasPrefix(v, "0x") || intV > 10000 {
						formattedValue = fmt.Sprintf("0x%x", intV)
					} else {
						formattedValue = fmt.Sprintf("%d", intV)
					}
				} else {
					formattedStart = "// " + formattedStart
					formattedValue = fmt.Sprintf("(unrecognized type) %v", v)
				}
			}
		case yaml.SequenceNode:
			// TODO
		default:
			return nil, fmt.Errorf("key %s has invalid value kind: %d", k, rv.Kind)
		}

		preset.Entries = append(preset.Entries, formattedStart+formattedValue)
	}
	return &preset, nil
}

func main() {
	var presetsDirPath, outputDirPath string
	flag.StringVar(&presetsDirPath, "presets-dir", "", "The file path to the directory containing yaml constant presets")
	flag.StringVar(&outputDirPath, "output-dir", "", "The file path to the directory to output generated Go code to")
	flag.Parse()

	templ := template.Must(template.New("preset").Parse(presetTemplate))

	if err := filepath.Walk(presetsDirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// skip directories
			if info.IsDir() {
				return nil
			}

			fmt.Println("processing preset file", path)

			extension := filepath.Ext(path)
			if extension != ".yaml" && extension != ".yml" {
				return nil
			}

			presetData, err := buildPreset(path)
			if err != nil {
				return err
			}

			outPath := filepath.Join(outputDirPath, presetData.Name+".go")
			fmt.Printf("writing constants preset %s to %s\n", presetData.Name, outPath)
			f, err := os.Create(outPath)
			if err != nil {
				return err
			}
			if err := templ.Execute(f, presetData); err != nil {
				return err
			}
			return nil
		}); err != nil {
		panic(err)
	}

}

var presetTemplate = `
package generated

{{ if .NeedsUtil }}
import "github.com/protolambda/zrnt/presets/util"
{{ end }}
const PRESET_NAME string = "{{.Name}}"
{{ range .Entries }}
{{.}}
{{ end }}`
