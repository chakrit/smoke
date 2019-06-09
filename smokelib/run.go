package smokelib

import (
	"bytes"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

func RunTests(tests []*Test) ([]*TestResult, error) {
	var results []*TestResult
	for _, test := range tests {
		if result, err := test.Run(); err != nil {
			return nil, err
		} else {
			results = append(results, result)
		}
	}
	return results, nil
}

func RunCommand(config *Config, c Command) (*Output, error) {
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
