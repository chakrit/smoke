package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/chakrit/smoke/checks"
	"github.com/pkg/errors"
)

func RunTest(t *Test) (TestResult, error) {
	var results []CommandResult
	for _, cmd := range t.Commands {
		if result, err := RunCommand(t.RunConfig, cmd, t.Checks); err != nil {
			return TestResult{}, err
		} else {
			results = append(results, result)
		}
	}

	return TestResult{
		Test:     t,
		Commands: results,
	}, nil
}

func RunCommand(config *Config, c Command, chks []checks.Interface) (CommandResult, error) {
	if config == nil {
		config = DefaultConfig
	}
	if config.Interpreter == "" {
		config.Interpreter = DefaultConfig.Interpreter
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig.Timeout
	}

	var (
		// -s causes most shell to read commands from the stdin
		// we use this approach to avoid having to argv parse by ourselves and get
		// closest to shell-native expectation in yaml files
		cmd   = exec.Command(config.Interpreter, "-s")
		errc  = make(chan error)
		inbuf = &bytes.Buffer{}
	)

	defer close(errc)
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	fmt.Fprintln(inbuf, string(c))
	cmd.Stdin = inbuf

	if err := checks.PrepareAll(cmd, chks); err != nil {
		return CommandResult{}, errors.Wrap(err, "checks")
	}
	if err := cmd.Start(); err != nil {
		return CommandResult{}, errors.Wrap(err, "start")
	}

	go func() { errc <- cmd.Wait() }()

	select {
	case <-time.After(config.Timeout):
		return CommandResult{
			Command: c,
			Err:     errors.New("timeout"),
		}, nil

	case err := <-errc: // Wait() returned
		if err == nil {
			// success case
		} else if _, ok := err.(*exec.ExitError); ok {
			// success case, with diff exit code
		} else {
			return CommandResult{}, errors.Wrap(err, "wait")
		}
	}

	if outputs, err := checks.CollectAll(cmd, chks); err != nil {
		return CommandResult{}, errors.Wrap(err, "checks")
	} else {
		return CommandResult{
			Command: c,
			Checks:  outputs,
		}, nil
	}
}
