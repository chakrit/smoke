package main

import (
	"bytes"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

type RunConfig struct {
	Env         []string      `yaml:"env"`
	Interpreter string        `yaml:"interpreter"`
	Timeout     time.Duration `yaml:"timeout"`
}

var defaultRunConfig = &RunConfig{
	Env:         nil,
	Interpreter: "/bin/bash",
	Timeout:     3 * time.Second,
}

func RunCommand(config *RunConfig, c Command) (*Output, error) {
	if config == nil {
		config = defaultRunConfig
	}
	if config.Interpreter == "" {
		config.Interpreter = defaultRunConfig.Interpreter
	}
	if config.Timeout == 0 {
		config.Timeout = defaultRunConfig.Timeout
	}

	var (
		// -s causes most shell to read commands from the stdin
		// we use this approach to avoid having to argv parse by ourselves and get
		// closest to shell-native expectation in yaml files
		cmd  = exec.Command(config.Interpreter, "-s")
		errc = make(chan error)

		inbuf  = &bytes.Buffer{}
		outbuf = &bytes.Buffer{}
		errbuf = &bytes.Buffer{}
	)

	defer close(errc)
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	inbuf.WriteString(string(c))
	inbuf.WriteByte(0x0D) // carriage-return
	cmd.Stdin = inbuf
	cmd.Stdout = outbuf
	cmd.Stderr = errbuf

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "start")
	}

	go func() { errc <- cmd.Wait() }()

	select {
	case <-time.After(config.Timeout):
		return nil, errors.New("timeout")

	case err := <-errc: // Wait() returned
		if _, ok := err.(*exec.ExitError); ok {
			// expected, nothing to do
		} else {
			return nil, errors.Wrap(err, "wait")
		}
	}

	return &Output{
		ExitCode: cmd.ProcessState.ExitCode(),
		Stdout:   outbuf.Bytes(),
		Stderr:   errbuf.Bytes(),
	}, nil
}
