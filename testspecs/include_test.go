package testspecs

import (
	"errors"
	"slices"
	"strings"
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

// Parameterized include: env flows down through Resolve, and the imported tests'
// names interpolate it via os.Expand — so one shared file, included under
// siblings that set different env, yields distinctly-named copies without editing
// the shared file.
func TestLoadIncludeParameterizedNames(t *testing.T) {
	tests, err := loadFiles(t, "root.yml", map[string]string{
		"root.yml": "tests:\n" +
			"  - name: postgres\n    config:\n      env: [\"DB=postgres\"]\n    include: db.yml\n" +
			"  - name: mysql\n    config:\n      env: [\"DB=mysql\"]\n    include: db.yml\n",
		"db.yml": "tests:\n  - name: \"connect-${DB}\"\n    commands: [\"echo hi\"]\n",
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	want := []string{
		`root.yml \ postgres \ db.yml \ connect-postgres`,
		`root.yml \ mysql \ db.yml \ connect-mysql`,
	}
	if got := names(tests); !slices.Equal(got, want) {
		t.Errorf("names = %v\nwant      %v", got, want)
	}
}

// An undefined variable expands to empty (the os.Expand/envsubst default), not an
// error — the source is the spec's declared env only, never os.Environ.
func TestLoadNameUndefinedVarExpandsEmpty(t *testing.T) {
	tests, err := loadString(t, "spec.yml",
		"tests:\n  - name: \"x-${NOPE}\"\n    commands: [\"echo hi\"]\n")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got, want := names(tests), []string{`spec.yml \ x-`}; !slices.Equal(got, want) {
		t.Errorf("names = %v, want %v", got, want)
	}
}

// A file that ends up among its own include ancestors is a cycle → exit 65.
func TestLoadIncludeCycleRejected(t *testing.T) {
	_, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "tests:\n  - name: toa\n    include: b.yml\n",
		"b.yml": "tests:\n  - name: tob\n    include: a.yml\n",
	})
	if err == nil {
		t.Fatal("want cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error should name the cycle, got: %v", err)
	}
	var se *SpecError
	if !errors.As(err, &se) {
		t.Errorf("cycle should be a SpecError (exit 65), got: %v", err)
	}
}

// A direct self-include is the degenerate cycle.
func TestLoadIncludeSelfCycleRejected(t *testing.T) {
	_, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "include: a.yml\n",
	})
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("want self-cycle error, got: %v", err)
	}
}

// A diamond (A includes B and C, both include D) is NOT a cycle: D is on neither
// ancestor stack, so it loads independently down each branch, distinct by parent.
func TestLoadIncludeDiamondAllowed(t *testing.T) {
	tests, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "tests:\n  - name: b\n    include: b.yml\n  - name: c\n    include: c.yml\n",
		"b.yml": "tests:\n  - name: tod\n    include: d.yml\n",
		"c.yml": "tests:\n  - name: tod\n    include: d.yml\n",
		"d.yml": "tests:\n  - name: leaf\n    commands: [\"echo d\"]\n",
	})
	if err != nil {
		t.Fatalf("diamond should load, got: %v", err)
	}
	want := []string{
		`a.yml \ b \ b.yml \ tod \ d.yml \ leaf`,
		`a.yml \ c \ c.yml \ tod \ d.yml \ leaf`,
	}
	if got := names(tests); !slices.Equal(got, want) {
		t.Errorf("names = %v\nwant      %v", got, want)
	}
}

// loadSpec dispatches each file by its own extension, so a .yml may include a
// .cue and a .jsonl — each decodes via its own loader and the trees merge.
func TestLoadIncludeCrossFormat(t *testing.T) {
	tests, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "tests:\n" +
			"  - name: fromcue\n    include: b.cue\n" +
			"  - name: fromjsonl\n    include: c.jsonl\n",
		"b.cue": "commands: [\"echo cue\"]\nchecks: [\"stdout\"]\n",
		"c.jsonl": `{"name":"L1","commands":["echo l1"]}` + "\n" +
			`{"name":"L2","commands":["echo l2"]}` + "\n",
	})
	if err != nil {
		t.Fatalf("cross-format include should load, got: %v", err)
	}
	want := []string{
		`a.yml \ fromcue \ b.cue`,
		`a.yml \ fromjsonl \ c.jsonl \ L1`,
		`a.yml \ fromjsonl \ c.jsonl \ L2`,
	}
	if got := names(tests); !slices.Equal(got, want) {
		t.Errorf("names = %v\nwant      %v", got, want)
	}
}

// A referenced file that doesn't exist is malformed content (the host named a
// missing file) → SpecError → exit 65, distinct from a missing *root* spec (2).
func TestLoadIncludeMissingFileRejected(t *testing.T) {
	_, err := loadFiles(t, "a.yml", map[string]string{
		"a.yml": "tests:\n  - name: host\n    include: nope.yml\n",
	})
	if err == nil {
		t.Fatal("want error for missing include, got nil")
	}
	var se *SpecError
	if !errors.As(err, &se) {
		t.Errorf("missing include should be a SpecError (exit 65), got: %v", err)
	}
}

func cmds(test *engine.Test) []string {
	out := make([]string, len(test.Commands))
	for i, c := range test.Commands {
		out[i] = string(c)
	}
	return out
}
