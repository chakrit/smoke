package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/chakrit/smoke/checks"
	"github.com/pkg/errors"
)

type Runner interface {
	Test(t *Test) (TestResult, error)
	Command(t *Test, cmd Command) (CommandResult, error)
}

type RunHooks interface {
	BeforeTest(t *Test)
	BeforeCommand(t *Test, cmd Command)
	AfterCommand(t *Test, cmd Command, result CommandResult, err error)
	AfterTest(t *Test, result TestResult, err error)
}

type DefaultRunner struct{ Hooks RunHooks }

func (r DefaultRunner) Test(t *Test) (result TestResult, err error) {
	if r.Hooks != nil {
		r.Hooks.BeforeTest(t)
		defer func() { r.Hooks.AfterTest(t, result, err) }()
	}

	var results []CommandResult
	for _, cmd := range t.Commands {
		if result, err := r.Command(t, cmd); err != nil {
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

func (r DefaultRunner) Command(t *Test, c Command) (result CommandResult, err error) {
	if r.Hooks != nil {
		r.Hooks.BeforeCommand(t, c)
		defer func() { r.Hooks.AfterCommand(t, c, result, err) }()
	}

	config := t.RunConfig
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

	if err := checks.PrepareAll(cmd, t.Checks); err != nil {
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

	if outputs, err := checks.CollectAll(cmd, t.Checks); err != nil {
		return CommandResult{}, errors.Wrap(err, "checks")
	} else {
		return CommandResult{
			Command: c,
			Checks:  outputs,
		}, nil
	}
}
