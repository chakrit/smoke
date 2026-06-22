package resultspecs

import (
	"fmt"
	"io"

	"github.com/chakrit/smoke/engine"
	"gopkg.in/yaml.v3"
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
		Name     engine.TestName     `yaml:"name"`
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
