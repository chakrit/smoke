package smokelib

import (
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
)
