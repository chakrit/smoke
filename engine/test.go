package engine

import (
	"log"

	"github.com/chakrit/smoke/checks"

	"github.com/pkg/errors"
)

type Test struct {
	Name      string             `yaml:"name"`
	RunConfig *Config            `yaml:"run"`
	Commands  []Command          `yaml:"commands"`
	Checks    []checks.Interface `yaml:"checks"`
}

func (t *Test) Run() (TestResult, error) {
	var results []CommandResult
	for _, cmd := range t.Commands {
		log.Println("test:", cmd)
		if output, err := RunCommand(t.RunConfig, cmd, t.Checks); err != nil {
			log.Println(errors.Wrap(err, "test"))
		} else {
			results = append(results, CommandResult{
				Command: cmd,
				Checks:  output,
			})
		}
	}

	return TestResult{
		Test:           t,
		CommandResults: results,
	}, nil
}
