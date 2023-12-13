package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/chakrit/smoke/checks"
	"golang.org/x/xerrors"
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
			_ = cmd.Process.Kill()
		}
	}()

	_, _ = fmt.Fprintln(inbuf, string(c))
	cmd.Stdin = inbuf
	if config.WorkDir != "" {
		cmd.Dir = config.WorkDir
	}
	if len(config.Env) > 0 {
		cmd.Env = append(cmd.Env, config.Env...)
	}

	if err := checks.PrepareAll(cmd, t.Checks); err != nil {
		return CommandResult{}, xerrors.Errorf("checks", err)
	}
	if err := cmd.Start(); err != nil {
		return CommandResult{}, xerrors.Errorf("start", err)
	}

	go func() { errc <- cmd.Wait() }()

	select {
	case <-time.After(config.Timeout):
		_ = cmd.Process.Kill()
		_ = <-errc // wait() should return by now (prevent send on close)
		return CommandResult{}, xerrors.New("timeout")

	case err = <-errc: // Wait() returned
		if err == nil {
			// success case
		} else if _, ok := err.(*exec.ExitError); ok {
			// success case, with diff exit code
		} else {
			return CommandResult{}, xerrors.Errorf("wait", err)
		}
	}

	if outputs, err := checks.CollectAll(cmd, t.Checks); err != nil {
		return CommandResult{}, xerrors.Errorf("checks", err)
	} else {
		return CommandResult{
			Command: c,
			Checks:  outputs,
		}, nil
	}
}
