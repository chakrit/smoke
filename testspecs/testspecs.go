package testspecs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chakrit/smoke/engine"
)

// SpecError marks a malformed-spec failure — a parse, validation, or bad-include
// error in a file SMOKE read but cannot use. The CLI maps it to exit 65
// (EX_DATAERR); an operational I/O failure reading the *root* spec is returned
// bare instead, so it maps to exit 2. See docs/spec/exit-codes.md.
type SpecError struct{ err error }

func (e *SpecError) Error() string { return e.err.Error() }
func (e *SpecError) Unwrap() error { return e.err }

func specErr(err error) error { return &SpecError{err} }

func Load(filename string) ([]*engine.Test, error) {
	root, err := loadSpec(filename, nil)
	if err != nil {
		return nil, err
	}

	root.Resolve(nil)

	tests, err := root.Tests()
	if err != nil {
		return nil, specErr(err)
	}
	if err := checkUniqueNames(tests); err != nil {
		return nil, specErr(err)
	}
	return tests, nil
}

// loadSpec reads and decodes one spec file into an unresolved tree. stack carries
// the recursion's ancestor paths so includes (added later) can guard cycles and
// classify a missing file by graph position; at the root it is nil.
func loadSpec(path string, stack []string) (*TestSpec, error) {
	loader, err := loaderFor(path)
	if err != nil {
		return nil, specErr(err)
	}

	file, err := os.Open(path)
	if err != nil {
		// The root spec failing to open is operational I/O trouble (exit 2); an
		// included spec failing to open is malformed content — the host named a
		// file that isn't there (exit 65). Stack depth tells the two apart.
		if len(stack) == 0 {
			return nil, err
		}
		return nil, specErr(fmt.Errorf("include %q: %w", path, err))
	}
	defer file.Close()

	root, err := loader.Load(file)
	if err != nil {
		return nil, specErr(err)
	}

	root.Filename = path
	if err := checkIncludeExclusive(root); err != nil {
		return nil, specErr(err)
	}
	return root, nil
}

// checkIncludeExclusive rejects a node that sets both `include` and `tests`: the
// splice can't both pull in a file and host inline children at one node. Runs
// before the splice, so the splice may assume the invariant holds.
func checkIncludeExclusive(node *TestSpec) error {
	if node.Include != "" && len(node.Children) > 0 {
		return fmt.Errorf("test %q: `include` and `tests` are mutually exclusive", node.Name)
	}
	for _, child := range node.Children {
		if err := checkIncludeExclusive(child); err != nil {
			return err
		}
	}
	return nil
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
