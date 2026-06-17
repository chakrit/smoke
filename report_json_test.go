package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/chakrit/smoke/resultspecs"
)

func TestStatusExitCode(t *testing.T) {
	cases := map[status]struct {
		label string
		code  int
	}{
		statusUnchanged: {"unchanged", 0},
		statusChanged:   {"changed", 1},
		statusNew:       {"new", 3},
	}

	for s, want := range cases {
		if got := s.String(); got != want.label {
			t.Errorf("%d.String() = %q, want %q", s, got, want.label)
		}
		if got := s.ExitCode(); got != want.code {
			t.Errorf("%d.ExitCode() = %d, want %d", s, got, want.code)
		}
	}
}

func TestJSONStatus(t *testing.T) {
	cases := map[resultspecs.Action]string{
		resultspecs.Equal:        "matched",
		resultspecs.Added:        "changed",
		resultspecs.InnerChanges: "changed",
		resultspecs.Removed:      "missing",
	}

	for action, want := range cases {
		if got := jsonStatus(action); got != want {
			t.Errorf("jsonStatus(%v) = %q, want %q", action, got, want)
		}
	}
}

func TestJSONReporter(t *testing.T) {
	edits := []resultspecs.TestEdit{{
		Name:   "Foo",
		Action: resultspecs.InnerChanges,
		Commands: []resultspecs.CommandEdit{{
			Name:   "echo hi",
			Action: resultspecs.InnerChanges,
			Checks: []resultspecs.CheckEdit{
				{Name: "stdout", Action: resultspecs.InnerChanges},
				{Name: "exitcode", Action: resultspecs.Equal},
			},
		}},
	}}

	var buf bytes.Buffer
	if err := (jsonReporter{w: &buf}).Report("test/foo.lock.yml", statusChanged, edits); err != nil {
		t.Fatalf("report: %v", err)
	}

	var got jsonReport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}

	if got.Status != "changed" || got.ExitCode != 1 {
		t.Errorf("top-level = %q/%d, want changed/1", got.Status, got.ExitCode)
	}
	if got.Lock != "test/foo.lock.yml" {
		t.Errorf("lock = %q", got.Lock)
	}
	if len(got.Tests) != 1 || got.Tests[0].Status != "changed" {
		t.Fatalf("tests = %+v", got.Tests)
	}

	checks := got.Tests[0].Commands[0].Checks
	if len(checks) != 2 || checks[0].Status != "changed" || checks[1].Status != "matched" {
		t.Errorf("checks = %+v", checks)
	}
}
