package testspecs

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chakrit/smoke/engine"
)

// loadString writes src to a temp file named filename (so the basename, and thus
// the root TestName, is filename) and loads it through the path-based Load.
func loadString(t *testing.T, filename, src string) ([]*engine.Test, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write spec %q: %v", filename, err)
	}
	return Load(path)
}

// loadFiles writes each name→content into one temp dir (subdir names supported)
// and loads root through the path-based Load, so includes resolve against real
// sibling files.
func loadFiles(t *testing.T, root string, files map[string]string) ([]*engine.Test, error) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %q: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %q: %v", name, err)
		}
	}
	return Load(filepath.Join(dir, root))
}

// names projects flattened tests to their string names for assertions.
func names(tests []*engine.Test) []string {
	out := make([]string, len(tests))
	for i, tc := range tests {
		out[i] = string(tc.Name)
	}
	return out
}

func findTest(tests []*engine.Test, name string) *engine.Test {
	for _, tc := range tests {
		if string(tc.Name) == name {
			return tc
		}
	}
	return nil
}

func TestLoadJSON(t *testing.T) {
	src := `{"config":{"interpreter":"/bin/sh"},` +
		`"tests":[{"name":"Echo","commands":["echo hi"],"checks":["stdout"]}]}`

	tests, err := loadString(t, "spec.json", src)
	if err != nil {
		t.Fatalf("load json: %v", err)
	}
	if len(tests) != 1 {
		t.Fatalf("want 1 test, got %d", len(tests))
	}
	if got := tests[0].Name; got != `spec.json \ Echo` {
		t.Errorf("name = %q", got)
	}
	if got := string(tests[0].Commands[0]); got != "echo hi" {
		t.Errorf("command = %q", got)
	}
}

func TestLoadJSONRejectsUnknownField(t *testing.T) {
	src := `{"chekcs":["stdout"],"commands":["echo hi"]}`

	_, err := loadString(t, "spec.json", src)
	if err == nil {
		t.Fatal("want error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "chekcs") {
		t.Errorf("error should name the offending field, got: %v", err)
	}
}

// DisallowUnknownFields recurses, so a typo nested under tests[] must also fail
// closed — not just top-level keys.
func TestLoadJSONRejectsNestedUnknownField(t *testing.T) {
	src := `{"tests":[{"name":"Echo","commands":["echo hi"],"chekcs":["stdout"]}]}`

	_, err := loadString(t, "spec.json", src)
	if err == nil {
		t.Fatal("want error for nested unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "chekcs") {
		t.Errorf("error should name the offending field, got: %v", err)
	}
}

func TestLoadJSONL(t *testing.T) {
	src := `{"name":"A","commands":["echo a"]}` + "\n" +
		`{"name":"B","commands":["echo b"]}` + "\n"

	tests, err := loadString(t, "spec.jsonl", src)
	if err != nil {
		t.Fatalf("load jsonl: %v", err)
	}
	if len(tests) != 2 {
		t.Fatalf("want 2 tests, got %d", len(tests))
	}
	if got := tests[0].Name; got != `spec.jsonl \ A` {
		t.Errorf("tests[0].Name = %q", got)
	}
	if got := tests[1].Name; got != `spec.jsonl \ B` {
		t.Errorf("tests[1].Name = %q", got)
	}
}

func TestLoadJSONLRejectsUnknownField(t *testing.T) {
	src := `{"name":"A","chekcs":["stdout"]}` + "\n"

	_, err := loadString(t, "spec.jsonl", src)
	if err == nil {
		t.Fatal("want error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "chekcs") {
		t.Errorf("error should name the offending field, got: %v", err)
	}
}

func TestLoadJSONC(t *testing.T) {
	src := `{
		// a line comment
		"config": {"interpreter": "/bin/sh"}, /* an inline block */
		/*
		   a multi-line
		   block comment
		*/
		"tests": [
			{"name": "Echo", "commands": ["echo hi"], "checks": ["stdout"]}
		]
	}`

	tests, err := loadString(t, "spec.jsonc", src)
	if err != nil {
		t.Fatalf("load jsonc: %v", err)
	}
	if len(tests) != 1 {
		t.Fatalf("want 1 test, got %d", len(tests))
	}
	if got := tests[0].Name; got != `spec.jsonc \ Echo` {
		t.Errorf("name = %q", got)
	}
	if got := string(tests[0].Commands[0]); got != "echo hi" {
		t.Errorf("command = %q", got)
	}
}

// The stripper is string-aware: comment markers inside string literals are data,
// not comments, and must survive untouched — including across escaped quotes.
func TestLoadJSONCPreservesCommentsInStrings(t *testing.T) {
	src := `{
		// a real comment
		"commands": [
			"echo a // not a comment",
			"echo /* also kept */ b",
			"echo \"q\" // still in string"
		],
		"checks": ["stdout"]
	}`

	tests, err := loadString(t, "spec.jsonc", src)
	if err != nil {
		t.Fatalf("load jsonc: %v", err)
	}

	want := []string{
		"echo a // not a comment",
		"echo /* also kept */ b",
		`echo "q" // still in string`,
	}
	for i, w := range want {
		if got := string(tests[0].Commands[i]); got != w {
			t.Errorf("command[%d] = %q, want %q", i, got, w)
		}
	}
}

// Stripping comments must not weaken the closed-decode boundary: a typo'd field
// still fails closed, same as plain .json.
func TestLoadJSONCRejectsUnknownField(t *testing.T) {
	src := `{
		// comment
		"chekcs": ["stdout"], "commands": ["echo hi"]
	}`

	_, err := loadString(t, "spec.jsonc", src)
	if err == nil {
		t.Fatal("want error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "chekcs") {
		t.Errorf("error should name the offending field, got: %v", err)
	}
}

// stripJSONComments must (1) preserve byte length so json.Decoder error offsets
// and line numbers still point at the original source, and (2) leave comment
// markers inside string literals untouched. Asserted through decode + the length
// invariant rather than brittle whitespace counting.
func TestStripJSONCommentsProperties(t *testing.T) {
	cases := []struct {
		name    string
		jsonc   string
		wantCmd string
	}{
		{"line and block stripped", "{\n// c\n\"commands\": [\"echo hi\"] /* x */\n}", "echo hi"},
		{"markers inside string kept", `{"commands": ["echo // /* keep */"]}`, "echo // /* keep */"},
		{"unterminated block after a complete value", `{"commands": ["a"]}/* tail`, "a"},
	}

	for _, c := range cases {
		stripped := stripJSONComments([]byte(c.jsonc))
		if len(stripped) != len(c.jsonc) {
			t.Errorf("%s: length not preserved (offset fidelity): in=%d out=%d",
				c.name, len(c.jsonc), len(stripped))
		}

		root := &TestSpec{}
		if err := decodeJSON(bytes.NewReader(stripped), root); err != nil {
			t.Fatalf("%s: decode: %v", c.name, err)
		}
		if got := root.Commands[0]; got != c.wantCmd {
			t.Errorf("%s: command = %q, want %q", c.name, got, c.wantCmd)
		}
	}
}

// A leaf (no children) with no commands is a malformed spec, not a silent
// skip — it must surface an error so the loader can exit 65.
func TestLoadLeafWithoutCommands(t *testing.T) {
	src := "tests:\n  - name: Empty\n"

	if _, err := loadString(t, "spec.yml", src); err == nil {
		t.Fatal("want error for command-less leaf, got nil")
	}
}

// Two tests that flatten to the same name make the lock ambiguous — a lock entry
// can't be resolved to one of them — so the loader rejects it as a malformed spec
// (exit 65), naming the duplicate.
func TestLoadRejectsDuplicateNames(t *testing.T) {
	src := "tests:\n" +
		"  - name: dup\n" +
		"    commands: [\"echo a\"]\n" +
		"  - name: dup\n" +
		"    commands: [\"echo b\"]\n"

	_, err := loadString(t, "spec.yml", src)
	if err == nil {
		t.Fatal("want error for duplicate test name, got nil")
	}
	if !strings.Contains(err.Error(), "dup") {
		t.Errorf("error should name the duplicate, got: %v", err)
	}
}

// include and tests are mutually exclusive on one node — both set is a malformed
// spec (exit 65): the splice would have to pull in a file *and* host inline
// children at the same node.
func TestLoadRejectsIncludeWithTests(t *testing.T) {
	src := "include: other.yml\n" +
		"tests:\n" +
		"  - name: inline\n" +
		"    commands: [\"echo hi\"]\n"

	_, err := loadString(t, "spec.yml", src)
	if err == nil {
		t.Fatal("want error for include+tests, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should explain the conflict, got: %v", err)
	}
}

// The root test name must be the spec's basename, not the path as typed on the
// command line — so the same spec yields the same TestName (and thus the same
// lock keys) regardless of directory depth or uncleaned `.`/`..` segments. For a
// single root spec the basename is its path "relative to the root spec file";
// imported specs extend that rule against the root's directory.
func TestLoadRootNameIsBasename(t *testing.T) {
	src := "tests:\n  - name: Echo\n    commands: [\"echo hi\"]\n"
	sub := filepath.Join(t.TempDir(), "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sub, "x.yml")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	// Every form points at the same file; only the basename may leak into the name.
	for _, form := range []string{path, sub + "/./x.yml", sub + "/../sub/x.yml"} {
		tests, err := Load(form)
		if err != nil {
			t.Fatalf("load %q: %v", form, err)
		}
		if got := tests[0].Name; got != `x.yml \ Echo` {
			t.Errorf("form %q: name = %q, want %q", form, got, `x.yml \ Echo`)
		}
	}
}

func TestLoadUnsupportedFormat(t *testing.T) {
	if _, err := loadString(t, "spec.txt", ""); err == nil {
		t.Fatal("want error for unsupported format, got nil")
	}
}

func TestLoadUnknownCheck(t *testing.T) {
	src := "commands: [\"echo hi\"]\nchecks: [\"nope://x\"]\n"

	_, err := loadString(t, "spec.yml", src)
	if err == nil {
		t.Fatal("want error for unknown check, got nil")
	}
	if !strings.Contains(err.Error(), "nope://x") {
		t.Errorf("error should name the check, got: %v", err)
	}
}

// A bad timeout duration is a data error (exit 65 territory).
func TestLoadBadTimeout(t *testing.T) {
	src := "commands: [\"echo hi\"]\nconfig:\n  timeout: \"nope\"\n"

	_, err := loadString(t, "spec.yml", src)
	if err == nil {
		t.Fatal("want error for bad timeout, got nil")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error should name the bad duration, got: %v", err)
	}
}

// Every independent tree-walk error surfaces from a single Load, not just the
// first — so a spec with several authoring mistakes is fixable in one pass. The
// uniqueness pass is a separate phase (still first-dup); this covers the flatten
// walk: a bad check, a bad timeout, and a command-less leaf across three nodes.
func TestLoadCollectsAllErrors(t *testing.T) {
	src := "tests:\n" +
		"  - name: BadCheck\n" +
		"    commands: [\"echo hi\"]\n" +
		"    checks: [\"boguscheck://x\"]\n" +
		"  - name: BadTimeout\n" +
		"    commands: [\"echo hi\"]\n" +
		"    config:\n" +
		"      timeout: \"notaduration\"\n" +
		"  - name: EmptyLeaf\n"

	_, err := loadString(t, "spec.yml", src)
	if err == nil {
		t.Fatal("want error for the malformed spec, got nil")
	}
	for _, want := range []string{"boguscheck://x", "notaduration", "EmptyLeaf"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q; got: %v", want, err)
		}
	}
}

// Every unknown check in one node surfaces, not just the first — the check loop
// accumulates rather than bailing on the first miss.
func TestLoadCollectsAllUnknownChecks(t *testing.T) {
	src := "commands: [\"echo hi\"]\nchecks: [\"bad1://x\", \"bad2://y\"]\n"

	_, err := loadString(t, "spec.yml", src)
	if err == nil {
		t.Fatal("want error for the unknown checks, got nil")
	}
	for _, want := range []string{"bad1://x", "bad2://y"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q; got: %v", want, err)
		}
	}
}

// A typo'd field on an otherwise-valid CUE node (here `chekcs` for `checks`)
// must fail closed against the schema, not silently drop the field at Decode.
func TestLoadCUERejectsUnknownField(t *testing.T) {
	src := "commands: [\"echo hi\"]\nchekcs: [\"stdout\"]\n"

	_, err := loadString(t, "spec.cue", src)
	if err == nil {
		t.Fatal("want error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "chekcs") {
		t.Errorf("error should name the offending field, got: %v", err)
	}
}

func TestLoadCUEValid(t *testing.T) {
	src := "commands: [\"echo hi\"]\nchecks: [\"stdout\"]\n"

	tests, err := loadString(t, "spec.cue", src)
	if err != nil {
		t.Fatalf("load cue: %v", err)
	}
	if len(tests) != 1 {
		t.Fatalf("want 1 test, got %d", len(tests))
	}
}
