package checks

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrCheck = errors.New("check problem")

	index = map[string]Interface{}
	all   = []Interface{
		StdoutCheck,
		StderrCheck,
		ExitCodeCheck,
	}
)

type (
	Interface interface {
		Name() string
		Prepare(cmd *exec.Cmd) error
		Collect(cmd *exec.Cmd) ([]byte, error)
	}

	Output struct {
		Name string
		Data []byte
	}
)

func init() {
	for _, check := range all {
		index[check.Name()] = check
	}
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

func Parse(line string) Interface {
	// TODO: more complex file system checks
	line = strings.TrimSpace(line)
	return index[line]
}

func PrepareAll(cmd *exec.Cmd, chks []Interface) error {
	for _, chk := range chks {
		if err := chk.Prepare(cmd); err != nil {
			return err
		}
	}
	return nil
}

func CollectAll(cmd *exec.Cmd, chks []Interface) ([]Output, error) {
	var outputs []Output
	for _, chk := range chks {
		if data, err := chk.Collect(cmd); err != nil {
			return nil, err
		} else {
			outputs = append(outputs, Output{
				Name: chk.Name(),
				Data: data,
			})
		}
	}
	return outputs, nil
}
