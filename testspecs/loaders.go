package testspecs

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"gopkg.in/yaml.v3"
)

// A loader parses one spec format into an unresolved root TestSpec. Format
// dispatch is default-deny: an unrecognized extension is rejected, never
// guessed.
type loader interface {
	Load(io.Reader) (*TestSpec, error)
}

func loaderFor(filename string) (loader, error) {
	switch ext := filepath.Ext(filename); ext {
	case ".yml", ".yaml", "":
		return yamlLoader{}, nil
	case ".cue":
		return cueLoader{}, nil
	case ".json":
		return jsonLoader{}, nil
	case ".jsonc":
		return jsoncLoader{}, nil
	case ".jsonl":
		return jsonlLoader{}, nil
	default:
		return nil, fmt.Errorf("unsupported spec format %q", ext)
	}
}

type yamlLoader struct{}

func (yamlLoader) Load(reader io.Reader) (*TestSpec, error) {
	root := &TestSpec{}
	if err := yaml.NewDecoder(reader).Decode(root); err != nil {
		return nil, err
	}
	return root, nil
}

//go:embed schema.cue
var cueSchema string

// cueLoader evaluates a single CUE file into the test tree, unifying it against
// the embedded `#Test` schema first so typo'd fields and wrong types fail closed
// as CUE constraint errors instead of being silently dropped at Decode. The
// struct's json tags drive the mapping; CUE has no native duration, so timeout
// stays a string parsed later in ConfigSpec.RunConfig (same path as YAML).
type cueLoader struct{}

func (cueLoader) Load(reader io.Reader) (*TestSpec, error) {
	src, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	ctx := cuecontext.New()
	schema := ctx.CompileString(cueSchema)
	if err := schema.Err(); err != nil {
		return nil, err
	}
	testDef := schema.LookupPath(cue.ParsePath("#Test"))

	value := ctx.CompileBytes(src)
	if err := value.Err(); err != nil {
		return nil, err
	}

	unified := value.Unify(testDef)
	if err := unified.Validate(cue.Concrete(false)); err != nil {
		return nil, err
	}

	root := &TestSpec{}
	if err := unified.Decode(root); err != nil {
		return nil, err
	}
	return root, nil
}

// decodeJSON decodes one JSON value into v, failing closed on unknown fields so a
// typo'd key (chekcs:) is rejected, not silently dropped — parity with the CUE
// loader's closed schema. Recurses through nested structs (config, tests[]).
func decodeJSON(reader io.Reader, v any) error {
	dec := json.NewDecoder(reader)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

type jsonLoader struct{}

func (jsonLoader) Load(reader io.Reader) (*TestSpec, error) {
	root := &TestSpec{}
	if err := decodeJSON(reader, root); err != nil {
		return nil, err
	}
	return root, nil
}

// jsoncLoader is the JSON loader with comment support: it strips // and /* */
// comments first, then decodes through the same closed path as plain JSON, so a
// typo'd field still fails closed. Comments only — trailing commas are not
// tolerated, same as .json.
type jsoncLoader struct{}

func (jsoncLoader) Load(reader io.Reader) (*TestSpec, error) {
	src, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	root := &TestSpec{}
	if err := decodeJSON(bytes.NewReader(stripJSONComments(src)), root); err != nil {
		return nil, err
	}
	return root, nil
}

// stripJSONComments blanks out // line and /* */ block comments, leaving string
// literals untouched (a // or /* inside a string is data, not a comment).
// Comment bytes become spaces and newlines are preserved, so the byte offsets
// and line numbers in json.Decoder errors still point at the original source.
func stripJSONComments(src []byte) []byte {
	const (
		normal = iota
		inString
		inStringEscape
		inLineComment
		inBlockComment
	)

	out := make([]byte, 0, len(src))
	state := normal

	for i := 0; i < len(src); i++ {
		c := src[i]
		switch state {
		case inString:
			out = append(out, c)
			switch c {
			case '\\':
				state = inStringEscape
			case '"':
				state = normal
			}

		case inStringEscape:
			out = append(out, c)
			state = inString

		case inLineComment:
			if c == '\n' {
				out = append(out, c)
				state = normal
			} else {
				out = append(out, ' ')
			}

		case inBlockComment:
			if c == '*' && peekByte(src, i) == '/' {
				out = append(out, ' ', ' ')
				state = normal
				i++
			} else if c == '\n' {
				out = append(out, c)
			} else {
				out = append(out, ' ')
			}

		default: // normal
			if c == '"' {
				out = append(out, c)
				state = inString
			} else if c == '/' && peekByte(src, i) == '/' {
				out = append(out, ' ', ' ')
				state = inLineComment
				i++
			} else if c == '/' && peekByte(src, i) == '*' {
				out = append(out, ' ', ' ')
				state = inBlockComment
				i++
			} else {
				out = append(out, c)
			}
		}
	}
	return out
}

func peekByte(src []byte, i int) byte {
	if i+1 < len(src) {
		return src[i+1]
	}
	return 0
}

// jsonlLoader reads one TestSpec per non-blank line. The stream is the children
// of an implicit empty root — equivalent to a YAML `tests: [...]` with no
// top-level command.
type jsonlLoader struct{}

func (jsonlLoader) Load(reader io.Reader) (*TestSpec, error) {
	root := &TestSpec{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		child := &TestSpec{}
		if err := decodeJSON(bytes.NewReader(line), child); err != nil {
			return nil, err
		}
		root.Children = append(root.Children, child)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return root, nil
}
