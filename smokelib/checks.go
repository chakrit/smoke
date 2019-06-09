package smokelib

import (
	"bytes"
	"os/exec"
	"strconv"
)

type checkImpl struct {
	name    string
	prepare func(cmd *exec.Cmd) error
	collect func(cmd *exec.Cmd) ([]byte, error)
}

var _ Check = &checkImpl{}

func (i *checkImpl) Name() string                          { return i.name }
func (i *checkImpl) Prepare(cmd *exec.Cmd) error           { return i.prepare(cmd) }
func (i *checkImpl) Collect(cmd *exec.Cmd) ([]byte, error) { return i.collect(cmd) }

var StdoutCheck = &checkImpl{
	name: "stdout",
	prepare: func(cmd *exec.Cmd) error {
		cmd.Stdout = &bytes.Buffer{}
		return nil
	},
	collect: func(cmd *exec.Cmd) ([]byte, error) {
		if buf, ok := cmd.Stdout.(*bytes.Buffer); !ok {
			return nil, ErrInternal
		} else {
			return buf.Bytes(), nil
		}
	},
}

var StderrCheck = &checkImpl{
	name: "stderr",
	prepare: func(cmd *exec.Cmd) error {
		cmd.Stderr = &bytes.Buffer{}
		return nil
	},
	collect: func(cmd *exec.Cmd) ([]byte, error) {
		if buf, ok := cmd.Stderr.(*bytes.Buffer); !ok {
			return nil, ErrInternal
		} else {
			return buf.Bytes(), nil
		}
	},
}

var ExitCodeCheck = &checkImpl{
	name: "exitcode",
	prepare: func(cmd *exec.Cmd) error {
		return nil
	},
	collect: func(cmd *exec.Cmd) ([]byte, error) {
		str := strconv.Itoa(cmd.ProcessState.ExitCode())
		return []byte(str), nil
	},
}
