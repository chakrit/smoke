package checks

import (
	"bytes"
	"os/exec"
	"strconv"
)

var ExitCodeCheck = &impl{
	name: "exitcode",
	prepare: func(cmd *exec.Cmd) error {
		return nil
	},
	collect: func(cmd *exec.Cmd) ([]byte, error) {
		str := strconv.Itoa(cmd.ProcessState.ExitCode())
		return []byte(str), nil
	},
}

var StdoutCheck = &impl{
	name: "stdout",
	prepare: func(cmd *exec.Cmd) error {
		cmd.Stdout = &bytes.Buffer{}
		return nil
	},
	collect: func(cmd *exec.Cmd) ([]byte, error) {
		if buf, ok := cmd.Stdout.(*bytes.Buffer); !ok {
			return nil, NewError("stdout", "not a buffer")
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
			return nil, NewError("stderr", "not a buffer")
		} else {
			return buf.Bytes(), nil
		}
	},
}
