package testspecs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

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

// cueLoader evaluates a single CUE file into the test tree. The struct's json
// tags drive the mapping; CUE has no native duration, so timeout stays a string
// parsed later in ConfigSpec.RunConfig (same path as YAML).
type cueLoader struct{}

func (cueLoader) Load(reader io.Reader) (*TestSpec, error) {
	src, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	value := cuecontext.New().CompileBytes(src)
	if err := value.Err(); err != nil {
		return nil, err
	}

	root := &TestSpec{}
	if err := value.Decode(root); err != nil {
		return nil, err
	}
	return root, nil
}

type jsonLoader struct{}

func (jsonLoader) Load(reader io.Reader) (*TestSpec, error) {
	root := &TestSpec{}
	if err := json.NewDecoder(reader).Decode(root); err != nil {
		return nil, err
	}
	return root, nil
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
		if err := json.Unmarshal(line, child); err != nil {
			return nil, err
		}
		root.Children = append(root.Children, child)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return root, nil
}
