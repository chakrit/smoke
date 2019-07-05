package engine

import (
	"github.com/chakrit/smoke/checks"
)

type (
	Command string

	Test struct {
		Name      string
		RunConfig *Config

		Commands []Command
		Checks   []checks.Interface
	}

	CommandResult struct {
		Command Command `yaml:"command"`
		Checks  []checks.Result
	}

	TestResult struct {
		Test     *Test `yaml:"test"`
		Commands []CommandResult
	}
)
