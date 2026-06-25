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
	"cuelang.org/go/cue/load"
	"gopkg.in/yaml.v3"
)

// A loader parses one spec format into an unresolved root TestSpec. Format
// dispatch is default-deny: an unrecognized extension is rejected, never
// guessed. path is the on-disk spec path: byte-stream loaders read from reader
// and ignore it; the CUE loader needs it to resolve cue.mod module imports.
type loader interface {
	Load(reader io.Reader, path string) (*TestSpec, error)
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

func (yamlLoader) Load(reader io.Reader, _ string) (*TestSpec, error) {
	root := &TestSpec{}
	if err := yaml.NewDecoder(reader).Decode(root); err != nil {
		return nil, err
	}
	return root, nil
}

//go:embed schema.cue
var cueSchema string

// cueLoader evaluates a CUE file into the test tree, unifying it against the
// embedded `#Test` schema first so typo'd fields and wrong types fail closed as
// CUE constraint errors instead of being silently dropped at Decode. The struct's
// json tags drive the mapping; CUE has no native duration, so timeout stays a
// string parsed later in ConfigSpec.RunConfig (same path as YAML).
//
// It loads by path (not the reader) via cue/load so a spec inside a cue.mod
// module can `import` shared packages; a lone .cue with no cue.mod loads just the
// same, as a single anonymous instance. The reader is unused — loadSpec already
// opened the file to classify a missing root (exit 2) vs include (exit 65) before
// we get here.
type cueLoader struct{}

func (cueLoader) Load(_ io.Reader, path string) (*TestSpec, error) {
	// cue/load resolves file args relative to Config.Dir, so a relative path plus
	// a Dir would double up (test/testdata/test/testdata/...). Absolutize first:
	// the arg is then anchored on its own, and Dir is the spec's directory where
	// cue.mod resolution begins.
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	insts := load.Instances([]string{abs}, &load.Config{Dir: filepath.Dir(abs)})
	if len(insts) != 1 {
		return nil, fmt.Errorf("cue: expected one instance for %q, got %d", path, len(insts))
	}
	if err := insts[0].Err; err != nil {
		return nil, err
	}

	ctx := cuecontext.New()
	schema := ctx.CompileString(cueSchema)
	if err := schema.Err(); err != nil {
		return nil, err
	}
	testDef := schema.LookupPath(cue.ParsePath("#Test"))

	value := ctx.BuildInstance(insts[0])
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

func (jsonLoader) Load(reader io.Reader, _ string) (*TestSpec, error) {
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

func (jsoncLoader) Load(reader io.Reader, _ string) (*TestSpec, error) {
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

func (jsonlLoader) Load(reader io.Reader, _ string) (*TestSpec, error) {
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
