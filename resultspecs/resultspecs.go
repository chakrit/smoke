package resultspecs

import (
	"bytes"
	"io"
	"strings"

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

func FromTestResult(result engine.TestResult) TestResultSpec {
	var commands []CommandResultSpec
	for _, cmd := range result.Commands {
		var checks []CheckOutputSpec
		for _, chk := range cmd.Checks {
			checks = append(checks, CheckOutputSpec{
				Name: chk.Name,
				Data: formatCheckOutput(chk.Data),
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
	}
}

func Load(r io.Reader) (specs []TestResultSpec, err error) {
	panic("not implemented")
}

func Save(w io.Writer, results []engine.TestResult) error {
	var specs []TestResultSpec
	for _, result := range results {
		specs = append(specs, FromTestResult(result))
	}

	if err := yaml.NewEncoder(w).Encode(specs); err != nil {
		return err
	}

	return nil
}

func formatCheckOutput(data []byte) []string {
	data = bytes.Trim(data, "\r\n")
	return strings.Split(string(data), "\n")
}
