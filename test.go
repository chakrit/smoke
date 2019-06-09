package main

import (
	"log"

	"github.com/pkg/errors"
)

type Test struct {
	Name      string     `yaml:"name"`
	WorkDir   string     `yaml:"workdir"`
	RunConfig *RunConfig `yaml:"run"`

	Setups   []Command `yaml:"setups"`
	Commands []Command `yaml:"commands"`
	Subtests []*Test   `yaml:"subtests"`
}

func (t *Test) Run() (*TestResult, error) {
	for _, setup := range t.Setups {
		log.Println("setup:", setup)
		_, err := RunCommand(t.RunConfig, setup)
		if err != nil {
			// no point to run tests if setup fails
			return nil, errors.Wrap(err, "setup")
		}
	}

	outputs := []*Output{}
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
