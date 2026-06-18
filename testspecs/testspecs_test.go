package testspecs

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoadJSON(t *testing.T) {
	src := `{"config":{"interpreter":"/bin/sh"},` +
		`"tests":[{"name":"Echo","commands":["echo hi"],"checks":["stdout"]}]}`

	tests, err := Load(strings.NewReader(src), "spec.json")
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

	_, err := Load(strings.NewReader(src), "spec.json")
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

	_, err := Load(strings.NewReader(src), "spec.json")
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

	tests, err := Load(strings.NewReader(src), "spec.jsonl")
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

	_, err := Load(strings.NewReader(src), "spec.jsonl")
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

	tests, err := Load(strings.NewReader(src), "spec.jsonc")
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

	tests, err := Load(strings.NewReader(src), "spec.jsonc")
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

	_, err := Load(strings.NewReader(src), "spec.jsonc")
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

	if _, err := Load(strings.NewReader(src), "spec.yml"); err == nil {
		t.Fatal("want error for command-less leaf, got nil")
	}
}

// Test identity is the flattened name; two siblings sharing a name collide on
// the same TestID. That ambiguity must be a load error — a name-keyed lock
// merge cannot tell the two apart, so the spec is rejected up front.
func TestLoadRejectsDuplicateNames(t *testing.T) {
	src := strings.Join([]string{
		"tests:",
		"  - name: Echo",
		"    commands: [\"echo a\"]",
		"  - name: Echo",
		"    commands: [\"echo b\"]",
	}, "\n")

	_, err := Load(strings.NewReader(src), "spec.yml")
	if err == nil {
		t.Fatal("want error for duplicate test name, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") || !strings.Contains(err.Error(), "Echo") {
		t.Errorf("error should flag the duplicate name, got: %v", err)
	}
}

func TestLoadUnsupportedFormat(t *testing.T) {
	if _, err := Load(strings.NewReader(""), "spec.txt"); err == nil {
		t.Fatal("want error for unsupported format, got nil")
	}
}

// All-errors reporting: a spec with three distinct mistakes in different
// subtrees — an unknown check, a bad timeout, and a command-less leaf — must
// surface ALL of them from one Load call, in depth-first spec order, so the
// author fixes everything in a single pass rather than fix-rerun-fix-rerun.
func TestLoadCollectsAllErrors(t *testing.T) {
	src := strings.Join([]string{
		"tests:",
		"  - name: BadCheck",
		"    commands: [\"echo a\"]",
		"    checks: [\"nope://x\"]",
		"  - name: BadTimeout",
		"    commands: [\"echo b\"]",
		"    config:",
		"      timeout: \"not-a-duration\"",
		"  - name: EmptyLeaf",
		"",
	}, "\n")

	_, err := Load(strings.NewReader(src), "spec.yml")
	if err == nil {
		t.Fatal("want aggregated error, got nil")
	}

	msg := err.Error()
	for _, want := range []string{"nope://x", "not-a-duration", "EmptyLeaf"} {
		if !strings.Contains(msg, want) {
			t.Errorf("aggregated error missing %q; got: %v", want, msg)
		}
	}

	// Stable depth-first order matching the spec: BadCheck → BadTimeout →
	// EmptyLeaf.
	iCheck := strings.Index(msg, "nope://x")
	iTimeout := strings.Index(msg, "not-a-duration")
	iLeaf := strings.Index(msg, "EmptyLeaf")
	if !(iCheck < iTimeout && iTimeout < iLeaf) {
		t.Errorf("errors out of spec order: check=%d timeout=%d leaf=%d in %q",
			iCheck, iTimeout, iLeaf, msg)
	}
}

// A single bad check still reports cleanly (single-error coverage retained
// through the fold).
func TestLoadUnknownCheck(t *testing.T) {
	src := "commands: [\"echo hi\"]\nchecks: [\"nope://x\"]\n"

	_, err := Load(strings.NewReader(src), "spec.yml")
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

	_, err := Load(strings.NewReader(src), "spec.yml")
	if err == nil {
		t.Fatal("want error for bad timeout, got nil")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error should name the bad duration, got: %v", err)
	}
}

// A typo'd field on an otherwise-valid CUE node (here `chekcs` for `checks`)
// must fail closed against the schema, not silently drop the field at Decode.
func TestLoadCUERejectsUnknownField(t *testing.T) {
	src := "commands: [\"echo hi\"]\nchekcs: [\"stdout\"]\n"

	_, err := Load(strings.NewReader(src), "spec.cue")
	if err == nil {
		t.Fatal("want error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "chekcs") {
		t.Errorf("error should name the offending field, got: %v", err)
	}
}

func TestLoadCUEValid(t *testing.T) {
	src := "commands: [\"echo hi\"]\nchecks: [\"stdout\"]\n"

	tests, err := Load(strings.NewReader(src), "spec.cue")
	if err != nil {
		t.Fatalf("load cue: %v", err)
	}
	if len(tests) != 1 {
		t.Fatalf("want 1 test, got %d", len(tests))
	}
}
