package resultspecs

import (
	"strings"
	"testing"
)

func spec(name string, cmds ...string) TestResultSpec {
	commands := make([]CommandResultSpec, len(cmds))
	for i, c := range cmds {
		commands[i] = CommandResultSpec{Command: c}
	}
	return TestResultSpec{Name: name, Commands: commands}
}

func names(specs []TestResultSpec) string {
	out := make([]string, len(specs))
	for i, s := range specs {
		out[i] = s.Name
	}
	return strings.Join(out, ",")
}

// Merge replaces base entries by identity, preserves unmatched base entries in
// place, and appends genuinely new overlay entries — the spine of partial
// commit, where a filtered run must not drop the tests it didn't run.
func TestMerge(t *testing.T) {
	tests := []struct {
		name    string
		base    []TestResultSpec
		overlay []TestResultSpec
		want    string
	}{
		{"empty overlay preserves base", []TestResultSpec{spec("A"), spec("B")}, nil, "A,B"},
		{"empty base takes overlay", nil, []TestResultSpec{spec("A"), spec("B")}, "A,B"},
		{"replace in place", []TestResultSpec{spec("A"), spec("B"), spec("C")}, []TestResultSpec{spec("B")}, "A,B,C"},
		{"append new after base", []TestResultSpec{spec("A")}, []TestResultSpec{spec("B")}, "A,B"},
		{"mixed replace and append", []TestResultSpec{spec("A"), spec("B")}, []TestResultSpec{spec("B"), spec("C")}, "A,B,C"},
		{"base order is stable", []TestResultSpec{spec("A"), spec("B"), spec("C")}, []TestResultSpec{spec("C"), spec("A")}, "A,B,C"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := names(Merge(tc.base, tc.overlay)); got != tc.want {
				t.Errorf("Merge order = %q, want %q", got, tc.want)
			}
		})
	}
}

// A matched entry must carry the overlay's observed output, not the base's —
// the whole point of committing is to replace the golden for that test.
func TestMergeReplacesContent(t *testing.T) {
	base := []TestResultSpec{spec("A", "old")}
	overlay := []TestResultSpec{spec("A", "new")}

	merged := Merge(base, overlay)
	if len(merged) != 1 {
		t.Fatalf("want 1 entry, got %d", len(merged))
	}
	if got := merged[0].Commands[0].Command; got != "new" {
		t.Errorf("merged command = %q, want overlay's %q", got, "new")
	}
}
