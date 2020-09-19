package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/docker/docker/daemon/logger/templates"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func mustTempl(tmpl string, data interface{}) string {
	t, err := templates.NewParse("", tmpl)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

func GoName(ssz SSZType) string {
	if ssz != nil {
		// basic types don't use pointers
		switch ssz.(type) {
		case *BoolDef, *RootDef, *BytesNDef, *UintNDef:
			return ssz.Name()
		default:
			// pointer
			return "*"+ssz.Name()
		}
	} else {
		panic("no ssz data")
	}
}

type ListDef struct {
	Base
	Elem string `yaml:"elem"`
	ElemSSZ SSZType
	Limit string `yaml:"limit"`
}

func (d *ListDef) GoElemName() string {
	return GoName(d.ElemSSZ, d.Elem)
}

func (d *ListDef) FmtLimit() string {
	// TODO format
	return d.Limit
}

type VectorDef struct {
	Base
	Elem string `yaml:"elem"`
	ElemSSZ SSZType
	Length string `yaml:"length"`
}

type FieldDef struct {
	Comment string
	Name string
	Type string
	// may be nil if it's not a custom type
	TypeSSZ SSZType
}

type ContainerDef struct {
	Base
	Fields []*FieldDef
}

type BitvectorDef struct {
	Base
	Length string `yaml:"length"`
}

type BitlistDef struct {
	Base
	Limit string `yaml:"limit"`
}

type AliasDef struct {
	Base
	Aliased string
	AliasedType SSZType
}

type BytesNDef struct {
	Base
	N uint
}

type UintNDef struct {
	Base
	N uint
}

type BoolDef struct {
	Base
}

type RootDef struct {
	Base
}



type SSZType interface {
	Name() string
	Comment() string
	Enhance()
	Build() string
	Basic() bool
	FixedLength() uint64
}

// TODO

type Base struct {
	Name string
	Comment string
	enhanced sync.Once
}

func (d *ListDef) Build() string {
	return mustTempl(`
	type {{.Name}} struct {
		Elements []{{.List.GoElemName}}
		Limit uint64
	}

	func (p *{{.Name}}) Deserialize(dr *codec.DecodingReader) error {
		// TODO
	}
	
	func (b *{{.Name}}) FixedLength() uint64 {
		// lists are never fixed length
		return 0
	}
	
	
	func (p *{{.Name}}) HashTreeRoot(hFn tree.HashFn) Root {
		
	}

	func (p *{{.Name}}) Serialize(w *EncodingWriter) error {

	}

	func (p *{{.Name}}) ValueByteLength() (out uint64) {
		for _, el := range p.Elements {
			out += el.ValueByteLength()
		}
		return
	}

	func (spec *Spec) CommitteeIndices() ListTypeDef {
		return ListType(spec.{{.List.Elem}}Type(), {{.List.FmtLimit}})
	}

	type {{.Name}}View struct{ *ComplexListView }  // TODO
	
	func As{{.Name}}(v View, err error) (*{{.Name}}View, error) {
		c, err := AsComplexList(v, err)
		return &{{.Name}}View{c}, nil
	}

	`, d)
}

var fieldNameRegex = regexp.MustCompile("^[A-Z][a-zA-Z0-9_]*$")

func checkFieldName(v string) error {
	if !fieldNameRegex.MatchString(v) {
		return fmt.Errorf("invalid field name: '%s'", v)
	}
	return nil
}

var typeNameRegex = regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$")

func checkTypeName(v string) error {
	if !typeNameRegex.MatchString(v) {
		return fmt.Errorf("invalid type name: '%s'", v)
	}
	return nil
}

type Entry struct{
	Def SSZType
}

func (s *Entry) UnmarshalYAML(root *yaml.Node) error {
	comm := ""
	if root.HeadComment != "" {
		comm += root.HeadComment
	}
	if root.LineComment != "" {
		if comm != "" {
			comm += "\n\n"
		}
		comm += root.LineComment
	}
	b := Base{Comment: comm}
	if root.Kind == yaml.ScalarNode {
		value := strings.TrimSpace(root.Value)
		if strings.HasPrefix("Bytes", value) {
			n, err := strconv.ParseUint(value[len("Bytes"):], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid BytesN: '%s'", value)
			}
			if !(n == 4 || n == 8 || n == 32 || n == 48 || n == 96) {
				return fmt.Errorf("unexpected BytesN: '%s'", value)
			}
			s.Def = &BytesNDef{Base: b, N: n}
			return nil
		}
		if strings.HasPrefix("uint", value) {
			n, err := strconv.ParseUint(value[len("uint"):], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid uintN: '%s'", value)
			}
			if !(n == 8 || n == 16 || n == 32 || n == 64) {
				return fmt.Errorf("unexpected uintN: '%s'", value)
			}
			s.Def = &UintNDef{Base: b, N: n}
			return nil
		}
		if value == "Root" {
			s.Def = &RootDef{Base: b}
			return nil
		}
		if value == "bool" {
			s.Def = &BoolDef{Base: b}
			return nil
		}
		if err := checkTypeName(value); err != nil {
			return err
		}
		s.Def = &AliasDef{Base: b, Aliased: value}
		return nil
	}

	if len(root.Content) != 2 {
		return fmt.Errorf("unexpected contents data, got %d kv entries", len(root.Content))
	}

	k := root.Content[0]
	if k.Kind != yaml.ScalarNode {
		return fmt.Errorf("unexpected key type: %v", k.Kind)
	}
	v := root.Content[1]
	switch k.Value {
	case "list":
		s.Def = &ListDef{Base: b}
	case "vector":
		s.Def = &VectorDef{Base: b}
	case "container":
		s.Def = &ContainerDef{Base: b}
	case "bitlist":
		s.Def = &BitlistDef{Base: b}
	case "bitvector":
		s.Def = &BitvectorDef{Base: b}
	default:
		return fmt.Errorf("unrecognized type: %s", k.Value)
	}
	if err := v.Decode(s.Def); err != nil {
		return err
	}
	return nil
}

type PhaseSchema struct {
	Types map[string]*Entry
}

func (s *PhaseSchema) UnmarshalYAML(root *yaml.Node) error {
	if err := root.Decode(&s.Types); err != nil {
		return err
	}
	for k, v := range s.Types {
		if err := checkTypeName(k); err != nil {
			return err
		}
		v.Name = k
		switch x := v.Type.(type) {
		case *AliasDef:
			if ref, ok := s.Types[x.Aliased.Name()]; ok {
				v.RefSSZ = ref
			} else {
				return fmt.Errorf("type '%s' references unknown type '%s'", k, v.Ref)
			}
		case *ContainerDef:
			for fi, fv := range v.Container.Fields {
				if ref, ok := s.Types[fv.Name]; ok {
					fv.TypeSSZ = ref
				} else {
					return fmt.Errorf("type '%s' field '%s' (index %d) references unknown type '%s'", k, fv.Name, fi, v.Ref)
				}
			}
		case *ListDef:
			if ref, ok := s.Types[v.List.Elem]; ok {
				v.List.ElemSSZ = ref
			} else {
				return fmt.Errorf("list type '%s' references unknown element type '%s'", k, v.List.Elem)
			}
		case *VectorDef:
			if ref, ok := s.Types[v.Vector.Elem]; ok {
				v.Vector.ElemSSZ = ref
			} else {
				return fmt.Errorf("vector type '%s' references unknown element type '%s'", k, v.Vector.Elem)
			}
		}
	}
	return nil
}

type Schema map[string]*PhaseSchema

func (s *Schema) Build() (string, error) {

}

func loadSchema(schemaPath string) (*Schema, error) {
	var schema Schema
	f, err := os.Open(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("config standard, cannot oppen file: %v", err)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

func main() {
	schemaPath := flag.String("schema", "schema.yaml", "schema to generate types for")
	typesOutputPath := flag.String("output", "types.go", "output go file to generate")
	//cfgStdPath := flag.String("config-spec", "config_spec.yaml", "configuration standard file to validate configs with")
	//forks := flag.String("forks", "phase0", "forks to expect, comma separated")
	flag.Parse()
	schema, err := loadSchema(*schemaPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load schema '%s': %v\n", *schemaPath, err)
		os.Exit(1)
	}
	out, err := schema.Build()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to generate code of schema '%s': %v\n", *schemaPath, err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(*typesOutputPath, []byte(out), 0664); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to write code of schema '%s': %v\n", *schemaPath, err)
		os.Exit(1)
	}
}
