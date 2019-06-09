package smokelib

import (
	"os/exec"

	"github.com/pkg/errors"
)

var (
	ErrInternal  = errors.New("internal integrity check failed")
	ErrSpec      = errors.New("bad spec")
	ErrCheckSpec = errors.New("bad check")
)

type (
	Command string

	Output struct {
		ExitCode int
		Stdout   []byte
		Stderr   []byte
	}

	TestResult struct {
		Test            *Test
		PreviousOutputs []*Output
		Outputs         []*Output
		Subresults      []*TestResult
	}

	Check interface {
		Name() string
		Prepare(cmd *exec.Cmd) error
		Collect(cmd *exec.Cmd) ([]byte, error)
	}
)
