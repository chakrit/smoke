package checks

import (
	"bytes"
	"os/exec"

	"github.com/pkg/errors"
)

var StdoutCheck = &impl{
	name: "stdout",
	prepare: func(cmd *exec.Cmd) error {
		cmd.Stdout = &bytes.Buffer{}
		return nil
	},
	collect: func(cmd *exec.Cmd) ([]byte, error) {
		if buf, ok := cmd.Stdout.(*bytes.Buffer); !ok {
			return nil, errors.WithMessage(ErrCheck, "stdout is not a buffer")
		} else {
			return buf.Bytes(), nil
		}
	},
}

var StderrCheck = &impl{
	name: "stderr",
	prepare: func(cmd *exec.Cmd) error {
		cmd.Stderr = &bytes.Buffer{}
		return nil
	},
	collect: func(cmd *exec.Cmd) ([]byte, error) {
		if buf, ok := cmd.Stderr.(*bytes.Buffer); !ok {
			return nil, errors.WithMessage(ErrCheck, "stderr is not a buffer")
		} else {
			return buf.Bytes(), nil
		}
	},
}
