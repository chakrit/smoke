package resultspecs

import (
	"fmt"
	"github.com/chakrit/smoke/engine"
	"gopkg.in/yaml.v3"
	"io"
)

const (
	NoOp = Action(iota)
	Equal
	Added
	Removed
	InnerChanges
)

type (
	Action int

	TestResultSpec struct {
		Name     string              `yaml:"name"`
		Commands []CommandResultSpec `yaml:"commands"`
	}
	CommandResultSpec struct {
		Command string            `yaml:"command"`
		Checks  []CheckOutputSpec `yaml:"checks"`
	}
	CheckOutputSpec struct {
		Name string   `yaml:"name"`
		Data []string `yaml:"data"`
	}
)

// ID is the test's identity for diffing and lock merges — derived from the
// stored name, so the on-disk format stays a bare `name`. Mirrors
// engine.Test's identity on the result side of the run/lock boundary.
func (s TestResultSpec) ID() engine.TestID { return engine.TestID(s.Name) }

func FromTestResult(result engine.TestResult) (TestResultSpec, error) {
	var commands []CommandResultSpec
	for _, cmd := range result.Commands {
		var checks []CheckOutputSpec
		for _, chk := range cmd.Checks {
			lines, err := chk.Check.Format(chk.Data)
			if err != nil {
				return TestResultSpec{}, fmt.Errorf("resultspecs: %w", err)
			}

			checks = append(checks, CheckOutputSpec{
				Name: chk.Check.Spec(),
				Data: lines,
			})
		}

		commands = append(commands, CommandResultSpec{
			Command: string(cmd.Command),
			Checks:  checks,
		})
	}

	return TestResultSpec{
		Name:     result.Test.Name,
		Commands: commands,
	}, nil
}

func Load(r io.Reader) (specs []TestResultSpec, err error) {
	if err := yaml.NewDecoder(r).Decode(&specs); err != nil {
		return nil, err
	} else {
		return specs, nil
	}
}

func FromTestResults(results []engine.TestResult) ([]TestResultSpec, error) {
	specs := make([]TestResultSpec, 0, len(results))
	for _, result := range results {
		spec, err := FromTestResult(result)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func Save(w io.Writer, specs []TestResultSpec) error {
	return yaml.NewEncoder(w).Encode(specs)
}

func Compare(oldspecs []TestResultSpec, newspecs []TestResultSpec) (edits []TestEdit, differs bool, err error) {
	return compareTests(oldspecs, newspecs)
}

// Merge overlays a (possibly partial) set of results onto a base lock by test
// identity: matched entries are replaced in their base position, unmatched base
// entries are preserved, and genuinely new entries append in overlay order.
func Merge(base, overlay []TestResultSpec) []TestResultSpec {
	replacement := make(map[engine.TestID]TestResultSpec, len(overlay))
	for _, spec := range overlay {
		replacement[spec.ID()] = spec
	}

	merged := make([]TestResultSpec, 0, len(base)+len(overlay))
	consumed := make(map[engine.TestID]bool, len(overlay))
	for _, spec := range base {
		if repl, ok := replacement[spec.ID()]; ok {
			merged = append(merged, repl)
			consumed[repl.ID()] = true
		} else {
			merged = append(merged, spec)
		}
	}

	for _, spec := range overlay {
		if !consumed[spec.ID()] {
			merged = append(merged, spec)
		}
	}
	return merged
}
