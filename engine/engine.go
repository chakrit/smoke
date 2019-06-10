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

	Test struct {
		Name      string
		RunConfig *Config

		Commands []Command
		Checks   []checks.Interface
	}
)
