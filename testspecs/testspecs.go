package testspecs

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/chakrit/smoke/engine"
)

func Load(reader io.Reader, filename string) ([]*engine.Test, error) {
	loader, err := loaderFor(filename)
	if err != nil {
		return nil, err
	}

	root, err := loader.Load(reader)
	if err != nil {
		return nil, err
	}

	root.Filename = filename
	root.Resolve(nil)

	tests, err := root.Tests()
	if err != nil {
		return nil, err
	}
	if err := checkUniqueNames(tests); err != nil {
		return nil, err
	}
	return tests, nil
}

// checkUniqueNames enforces that every flattened TestName is distinct. A
// collision makes the lock ambiguous — a name-keyed entry can't be resolved to
// one of two tests — so it is a malformed spec, surfaced for exit 65.
func checkUniqueNames(tests []*engine.Test) error {
	seen := make(map[engine.TestName]struct{}, len(tests))
	for _, test := range tests {
		if _, dup := seen[test.Name]; dup {
			return fmt.Errorf("duplicate test name `%s`", test.Name)
		}
		seen[test.Name] = struct{}{}
	}
	return nil
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
