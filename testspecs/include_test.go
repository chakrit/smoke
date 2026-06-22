package testspecs

import (
	"slices"
	"testing"

	"github.com/chakrit/smoke/engine"
)

// An include splices the referenced file's root in as a child of the importing
// node — two segment-bearing nodes (importing node, then imported root named by
// the include path), with the imported file's tests beneath.
func TestLoadIncludeSplicesAsChild(t *testing.T) {
	tests, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "tests:\n  - name: host\n    include: b.yml\n",
		"b.yml": "tests:\n  - name: greet\n    commands: [\"echo hi\"]\n",
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got, want := names(tests), []string{`a.yml \ host \ b.yml \ greet`}; !slices.Equal(got, want) {
		t.Errorf("names = %v, want %v", got, want)
	}
}

// A top-level include (on the root node itself) splices the imported root as a
// child of the root: `<root> \ <imported root> \ <imported tests>`.
func TestLoadIncludeAtRoot(t *testing.T) {
	tests, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "include: b.yml\n",
		"b.yml": "tests:\n  - name: greet\n    commands: [\"echo hi\"]\n",
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got, want := names(tests), []string{`a.yml \ b.yml \ greet`}; !slices.Equal(got, want) {
		t.Errorf("names = %v, want %v", got, want)
	}
}

// The imported root segment defaults to the include path as written, but an
// imported file that names its own root keeps that name (parity with the root
// basename default: an explicit name always wins).
func TestLoadIncludeRespectsImportedRootName(t *testing.T) {
	tests, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "tests:\n  - name: host\n    include: b.yml\n",
		"b.yml": "name: database\ntests:\n  - name: greet\n    commands: [\"echo hi\"]\n",
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got, want := names(tests), []string{`a.yml \ host \ database \ greet`}; !slices.Equal(got, want) {
		t.Errorf("names = %v, want %v", got, want)
	}
}

// The imported tree inherits from the importing node through the existing Resolve,
// identical to an inline child: config merges and commands prepend (parent before
// child). The importing node, carrying its own commands, also runs as its own test.
func TestLoadIncludeInheritsFromImportingNode(t *testing.T) {
	tests, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "tests:\n" +
			"  - name: host\n" +
			"    config:\n      env: [\"WHO=world\"]\n" +
			"    commands: [\"echo setup\"]\n" +
			"    include: b.yml\n",
		"b.yml": "tests:\n  - name: greet\n    commands: [\"echo hi\"]\n",
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	greet := findTest(tests, `a.yml \ host \ b.yml \ greet`)
	if greet == nil {
		t.Fatalf("imported test not found in %v", names(tests))
	}
	if got := cmds(greet); !slices.Equal(got, []string{"echo setup", "echo hi"}) {
		t.Errorf("commands = %v, want [echo setup, echo hi] (parent prepended)", got)
	}
	if !slices.Contains(greet.RunConfig.Env, "WHO=world") {
		t.Errorf("env = %v, want to inherit WHO=world", greet.RunConfig.Env)
	}
}

func cmds(test *engine.Test) []string {
	out := make([]string, len(test.Commands))
	for i, c := range test.Commands {
		out[i] = string(c)
	}
	return out
}
