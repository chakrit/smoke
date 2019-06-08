package main

import "github.com/pkg/errors"

type Test struct {
	Name        string   `yaml:"name"`
	WorkDir     string   `yaml:"workdir"`
	Interpreter string   `yaml:"interpreter"`
	Env         []string `yaml:"env"`

	Setups   []Command `yaml:"setups"`
	Commands []Command `yaml:"commands"`
	Subtests []*Test   `yaml:"tests"`
}

func (t *Test) Run() error {
	return errors.Wrap(errors.New("not implemented"),
		t.Name)
}
