package engine

import (
	"bytes"
	"os/exec"
	"time"

	"github.com/chakrit/smoke/checks"

	"github.com/pkg/errors"
)

func RunTests(tests []*Test) ([]TestResult, error) {
	var results []TestResult
	for _, test := range tests {
		if result, err := test.Run(); err != nil {
			return nil, err
		} else {
			results = append(results, result)
		}
	}
	return results, nil
}

func RunCommand(config *Config, c Command, chks []checks.Interface) ([]checks.Output, error) {
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

	inbuf.WriteString(string(c))
	inbuf.WriteByte(0x0D) // carriage-return
	cmd.Stdin = inbuf

	if err := checks.PrepareAll(cmd, chks); err != nil {
		return nil, errors.Wrap(err, "checks")
	}
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

	if outputs, err := checks.CollectAll(cmd, chks); err != nil {
		return nil, errors.Wrap(err, "checks")
	} else {
		return outputs, nil
	}
}
