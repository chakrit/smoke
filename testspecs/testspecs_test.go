package testspecs

import (
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

// A leaf (no children) with no commands is a malformed spec, not a silent
// skip — it must surface an error so the loader can exit 65.
func TestLoadLeafWithoutCommands(t *testing.T) {
	src := "tests:\n  - name: Empty\n"

	if _, err := Load(strings.NewReader(src), "spec.yml"); err == nil {
		t.Fatal("want error for command-less leaf, got nil")
	}
}

func TestLoadUnsupportedFormat(t *testing.T) {
	if _, err := Load(strings.NewReader(""), "spec.txt"); err == nil {
		t.Fatal("want error for unsupported format, got nil")
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
