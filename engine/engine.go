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
		Checks  []checks.Output
	}

	TestResult struct {
		Test           *Test
		CommandResults []CommandResult
	}
)
