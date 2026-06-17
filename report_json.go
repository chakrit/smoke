package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/chakrit/smoke/resultspecs"
)

type jsonReport struct {
	Status   string     `json:"status"`
	ExitCode int        `json:"exitCode"`
	Lock     string     `json:"lock"`
	Tests    []jsonTest `json:"tests"`
}

type jsonTest struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Commands []jsonCommand `json:"commands"`
}

type jsonCommand struct {
	Command string      `json:"command"`
	Status  string      `json:"status"`
	Checks  []jsonCheck `json:"checks"`
}

type jsonCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// jsonReporter renders the compare outcome as a machine-readable document — the
// agentic consumer surface. Detail localizes to drift: a node reported `matched`
// carries no children (a matched subtree needs no localization), while a
// `changed` node enumerates down to the per-check verdict. This mirrors the diff
// tree Compare produces — equal subtrees stay collapsed. Output is
// struct/slice-only (no maps), so field order is deterministic and the lock
// stays stable.
type jsonReporter struct {
	w io.Writer
}

func (r jsonReporter) Report(lock string, st status, edits []resultspecs.TestEdit) error {
	report := jsonReport{
		Status:   st.String(),
		ExitCode: st.ExitCode(),
		Lock:     lock,
		Tests:    jsonTests(edits),
	}

	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(r.w, string(encoded))
	return err
}

func jsonTests(edits []resultspecs.TestEdit) []jsonTest {
	tests := make([]jsonTest, 0, len(edits))
	for _, edit := range edits {
		tests = append(tests, jsonTest{
			Name:     edit.Name,
			Status:   jsonStatus(edit.Action),
			Commands: jsonCommands(edit.Commands),
		})
	}
	return tests
}

func jsonCommands(edits []resultspecs.CommandEdit) []jsonCommand {
	commands := make([]jsonCommand, 0, len(edits))
	for _, edit := range edits {
		commands = append(commands, jsonCommand{
			Command: edit.Name,
			Status:  jsonStatus(edit.Action),
			Checks:  jsonChecks(edit.Checks),
		})
	}
	return commands
}

func jsonChecks(edits []resultspecs.CheckEdit) []jsonCheck {
	checks := make([]jsonCheck, 0, len(edits))
	for _, edit := range edits {
		checks = append(checks, jsonCheck{
			Name:   edit.Name,
			Status: jsonStatus(edit.Action),
		})
	}
	return checks
}

// jsonStatus maps an edit Action to the per-node status vocabulary. `new` is a
// whole-run state (no lock at all), so it never appears at node level.
func jsonStatus(action resultspecs.Action) string {
	switch action {
	case resultspecs.Removed:
		return "missing"
	case resultspecs.Added, resultspecs.InnerChanges:
		return "changed"
	default: // Equal, NoOp
		return "matched"
	}
}
