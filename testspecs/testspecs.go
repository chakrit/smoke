package testspecs

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

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
	abs := cleanAbs(path)
	// stack is the ancestor chain, not a global visited-set: a file already among
	// its own ancestors is a true cycle (65); the same file reached down two
	// distinct branches (a diamond) is not, so it loads independently each time.
	if slices.Contains(stack, abs) {
		return nil, specErr(fmt.Errorf("include cycle: %q includes itself", path))
	}

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

	// Resolve includes relative to this file's directory (D2), pushing this file
	// onto the ancestor stack for the recursion. Clone so sibling include branches
	// (a diamond) never see each other's pushes.
	stack = append(slices.Clone(stack), abs)
	if err := spliceIncludes(root, filepath.Dir(path), stack); err != nil {
		return nil, err
	}
	return root, nil
}

// spliceIncludes walks node and its descendants and, for each `include`, loads
// the referenced file (relative to dir, D2) and splices its root in as a child
// (D3, two-node). The imported root's segment defaults to the include path as
// written when the file names no root of its own. Errors from the recursive load
// are already classified (SpecError), so they propagate as-is.
func spliceIncludes(node *TestSpec, dir string, stack []string) error {
	if node.Include != "" {
		childRoot, err := loadSpec(filepath.Join(dir, node.Include), stack)
		if err != nil {
			return err
		}
		if childRoot.Name == "" {
			childRoot.Name = node.Include
		}
		node.Children = append(node.Children, childRoot)
		node.Include = ""
		return nil
	}
	for _, child := range node.Children {
		if err := spliceIncludes(child, dir, stack); err != nil {
			return err
		}
	}
	return nil
}

// cleanAbs is the absolute, cleaned form of path — the stable key for the cycle
// guard, so the same file reached via ./x, x, or ../d/x collapses to one entry.
func cleanAbs(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return filepath.Clean(path)
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
