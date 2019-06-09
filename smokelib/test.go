package smokelib

import (
	"log"

	"github.com/pkg/errors"
)

type Test struct {
	Name      string    `yaml:"name"`
	RunConfig *Config   `yaml:"run"`
	Commands  []Command `yaml:"commands"`
	Checks    []Check   `yaml:"checks"`
}

func (t *Test) Run() (*TestResult, error) {
	var outputs []*Output
	for _, cmd := range t.Commands {
		log.Println("test:", cmd)
		if output, err := RunCommand(t.RunConfig, cmd); err != nil {
			log.Println(errors.Wrap(err, "test"))
		} else {
			outputs = append(outputs, output)
		}
	}

	return &TestResult{
		Test:       t,
		Outputs:    outputs,
		Subresults: nil,
	}, nil
}
