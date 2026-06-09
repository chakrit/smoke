package testspecs

import (
	"fmt"
	"io"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
	"github.com/chakrit/smoke/engine"
	"gopkg.in/yaml.v3"
)

func Load(reader io.Reader, filename string) ([]*engine.Test, error) {
	root := &TestSpec{}

	switch ext := filepath.Ext(filename); ext {
	case ".cue":
		if err := decodeCUE(reader, root); err != nil {
			return nil, err
		}
	case ".yml", ".yaml", "":
		if err := yaml.NewDecoder(reader).Decode(root); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported spec format %q", ext)
	}

	root.Filename = filename
	root.Resolve(nil)
	return root.Tests()
}

// decodeCUE evaluates a single CUE file into the test tree. The struct's json
// tags drive the mapping; CUE has no native duration, so timeout is a string
// parsed later in ConfigSpec.RunConfig (same path as YAML).
func decodeCUE(reader io.Reader, root *TestSpec) error {
	src, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	value := cuecontext.New().CompileBytes(src)
	if err := value.Err(); err != nil {
		return err
	}
	return value.Decode(root)
}

func resolvePaths(strs ...string) string {
	return filepath.Join(strs...)
}

func resolveStrings(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}
