package checks

import (
	"os/exec"
	"strings"
)

var (
	ExitCode Interface = exitCode{}
	Stdout   Interface = stdout{}
	Stderr   Interface = stderr{}
	// TODO: Binary versions
	// TODO: File-based checks

	index = map[string]Interface{}
	all   = []Interface{
		ExitCode,
		Stdout,
		Stderr,
	}
)

type (
	Interface interface {
		Name() string
		Prepare(*exec.Cmd) error
		Collect(*exec.Cmd) ([]byte, error)
		Format([]byte) ([]string, error)
	}

	Result struct {
		Check Interface
		Data  []byte
	}
)

func init() {
	for _, check := range all {
		index[check.Name()] = check
	}
}

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

func CollectAll(cmd *exec.Cmd, chks []Interface) ([]Result, error) {
	var results []Result
	for _, chk := range chks {
		if data, err := chk.Collect(cmd); err != nil {
			return nil, err
		} else {
			results = append(results, Result{
				Check: chk,
				Data:  data,
			})
		}
	}
	return results, nil
}
