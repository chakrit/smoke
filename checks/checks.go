package checks

import (
	"os/exec"

	"github.com/pkg/errors"
)

var (
	ErrCheck = errors.New("check problem")
)

type Interface interface {
	Name() string
	Prepare(cmd *exec.Cmd) error
	Collect(cmd *exec.Cmd) ([]byte, error)
}

type impl struct {
	name    string
	prepare func(cmd *exec.Cmd) error
	collect func(cmd *exec.Cmd) ([]byte, error)
}

var _ Interface = &impl{}

func (i *impl) Name() string                          { return i.name }
func (i *impl) Prepare(cmd *exec.Cmd) error           { return i.prepare(cmd) }
func (i *impl) Collect(cmd *exec.Cmd) ([]byte, error) { return i.collect(cmd) }

func ByName(name string) Interface {
	switch name {
	case "stdout":
		return StdoutCheck
	case "stderr":
		return StderrCheck
	case "exitcode":
		return ExitCodeCheck
	default:
		return nil
	}
}
