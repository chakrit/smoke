package resultspecs

import (
	"io"

	"golang.org/x/xerrors"

	"github.com/chakrit/smoke/engine"
	"gopkg.in/yaml.v3"
)

type (
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

func FromTestResult(result engine.TestResult) (TestResultSpec, error) {
	var commands []CommandResultSpec
	for _, cmd := range result.Commands {
		var checks []CheckOutputSpec
		for _, chk := range cmd.Checks {
			lines, err := chk.Check.Format(chk.Data)
			if err != nil {
				return TestResultSpec{}, xerrors.Errorf("resultspecs: %w", err)
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
