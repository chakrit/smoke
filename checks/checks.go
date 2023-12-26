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

// It maybe better to:
// 1. Follow the kubernetes style of `name: {}` or..
// 2. Uses polymorphic YAML or...
// 3. Writes a proper parser
func Parse(line string) Interface {
	lowline := strings.ToLower(line)
	if strings.HasPrefix(lowline, "get ") ||
		strings.HasPrefix(lowline, "head ") ||
		strings.HasPrefix(lowline, "put ") ||
		strings.HasPrefix(lowline, "post ") ||
		strings.HasPrefix(lowline, "patch ") ||
		strings.HasPrefix(lowline, "delete ") ||
		strings.HasPrefix(lowline, "options ") ||
		strings.HasPrefix(lowline, "trace ") {

		parts := strings.SplitN(line, " ", 2)
		return httpRequest{parts[0], parts[1]}
	}

	u, err := url.Parse(strings.TrimSpace(line))
	if err != nil {
		return nil
	}

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
