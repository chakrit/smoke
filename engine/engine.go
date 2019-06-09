package engine

import (
	"github.com/chakrit/smoke/checks"
	"github.com/pkg/errors"
)

var (
	ErrSpec = errors.New("bad spec")
)

type (
	Command string

	CommandResult struct {
		Command Command
		Err     error
		Checks  []checks.Output
	}

	Test struct {
		Name      string             `yaml:"name"`
		RunConfig *Config            `yaml:"run"`
		Commands  []Command          `yaml:"commands"`
		Checks    []checks.Interface `yaml:"checks"`
	}

	TestResult struct {
		Test     *Test
		Err      error
		Commands []CommandResult
	}
)
