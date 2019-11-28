package checks

import (
	"net/url"
	"os/exec"
	"strings"
)

var (
	ExitCode Interface = exitCode{}
	Stdout   Interface = stdout{}
	Stderr   Interface = stderr{}
	// TODO: Binary versions of STDIO
	// TODO: HTTP checks
)

type (
	Interface interface {
		Spec() string
		Prepare(*exec.Cmd) error
		Collect(*exec.Cmd) ([]byte, error)
		Format([]byte) ([]string, error)
	}

	Result struct {
		Check Interface
		Data  []byte
	}
)

func Parse(line string) Interface {
	u, err := url.Parse(strings.TrimSpace(line))
	if err != nil {
		return nil
	}

	// It maybe better to follow the kubernetes style of `name: {}` or just go
	// with polymorphic YAML, this feels magical and unidiomatic.
	switch u.Scheme {
	case "":
		switch strings.ToLower(strings.TrimSpace(u.Path)) {
		case "exitcode":
			return exitCode{}
		case "stdout":
			return stdout{}
		case "stderr":
			return stderr{}
		default:
			return fileContent{u.Path}
		}

	// TODO: case "http"
	default:
		return nil
	}
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
