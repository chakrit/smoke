package checks

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

type (
	stdout struct{}
	stderr struct{}
)

func (stdout) Spec() string { return "stdout" }
func (stderr) Spec() string { return "stderr" }
func (stdout) Name() string { return "stdout" }
func (stderr) Name() string { return "stderr" }

func (stdout) Prepare(cmd *exec.Cmd) error {
	cmd.Stdout = &bytes.Buffer{}
	return nil
}
func (stderr) Prepare(cmd *exec.Cmd) error {
	cmd.Stderr = &bytes.Buffer{}
	return nil
}

func (stdout) Collect(cmd *exec.Cmd) ([]byte, error) {
	if buf, ok := cmd.Stdout.(*bytes.Buffer); !ok {
		return nil, errors.New("stdout: pipe is not a bytes.Buffer{}")
	} else {
		return buf.Bytes(), nil
	}
}
func (stderr) Collect(cmd *exec.Cmd) ([]byte, error) {
	if buf, ok := cmd.Stderr.(*bytes.Buffer); !ok {
		return nil, errors.New("stderr: pipe is not a bytes.Buffer{}")
	} else {
		return buf.Bytes(), nil
	}
}

func (stderr) Format(buf []byte) ([]string, error) {
	str := string(bytes.Trim(buf, "\r\n"))
	return strings.Split(str, "\n"), nil
}
func (stdout) Format(buf []byte) ([]string, error) {
	str := string(bytes.Trim(buf, "\r\n"))
	return strings.Split(str, "\n"), nil
}
