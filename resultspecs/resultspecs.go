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

func Save(w io.Writer, results []engine.TestResult) error {
	var specs []TestResultSpec
	for _, result := range results {
		if spec, err := FromTestResult(result); err != nil {
			return err
		} else {
			specs = append(specs, spec)
		}
	}

	if err := yaml.NewEncoder(w).Encode(specs); err != nil {
		return err
	} else {
		return nil
	}
}

func Compare(oldspecs []TestResultSpec, newspecs []TestResultSpec) (edits []TestEdit, differs bool, err error) {
	return compareTests(oldspecs, newspecs)
}
