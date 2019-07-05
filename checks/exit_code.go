package checks

import (
	"os/exec"
	"strconv"
)

type exitCode struct{}

func (exitCode) Name() string {
	return "exitcode"
}

func (exitCode) Prepare(*exec.Cmd) error {
	return nil
}

func (exitCode) Collect(cmd *exec.Cmd) ([]byte, error) {
	str := strconv.Itoa(cmd.ProcessState.ExitCode())
	return []byte(str), nil
}

func (exitCode) Format(buf []byte) ([]string, error) {
	if len(buf) == 0 {
		return []string{}, nil
	} else {
		return []string{string(buf)}, nil
	}
}
