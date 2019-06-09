package checks

import (
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
