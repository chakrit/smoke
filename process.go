package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/chakrit/smoke/engine"
	"github.com/chakrit/smoke/internal/p"
	"github.com/chakrit/smoke/resultspecs"
	"github.com/chakrit/smoke/testspecs"
	"golang.org/x/xerrors"
)

func processFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		p.Exit(xerrors.Errorf(filename+": %w", err))
	}

	defer file.Close()

	tests, err := testspecs.Load(file, filename)
	if err != nil {
		p.Exit(xerrors.Errorf(filename+": %w", err))
	}

	if shouldList {
		listTests(tests)
		return
	}

	p.Action("Running Tests")
	results := runTests(tests)
	if shouldCommit {
		p.Action("Writing Lock File")
		commitResults(filename, results)
	} else {
		p.Action("Comparing Lock File")
		compareResults(filename, results)
	}
}

func listTests(tests []*engine.Test) {
	for _, test := range tests {
		p.Test(test)
		for _, cmd := range test.Commands {
			p.Command(test, cmd)
		}
	}
}

func runTests(tests []*engine.Test) []engine.TestResult {
	var (
		failed = false
		hooks  = Hooks{
			WrapErr: func(t *engine.Test, err error) error {
				return xerrors.Errorf(t.Name+": %w", err)
			},
		}

		run     engine.Runner = engine.DefaultRunner{Hooks: hooks}
		results []engine.TestResult
	)

	for _, test := range tests {
		if result, err := run.Test(test); err != nil {
			// error already printed by Hooks{}
			failed = true
			break
		} else {
			results = append(results, result)
		}
	}

	if failed {
		// TODO: Also fail, if testresult.any(:failed?)
		os.Exit(1)
		return nil
	}

	return results
}

func commitResults(filename string, results []engine.TestResult) {
	tmpfile, err := ioutil.TempFile("", "smoke-*.yml")
	if err != nil {
		p.Exit(xerrors.Errorf("commit: %w", err))
	}

	defer tmpfile.Close()

	p.FileAccess(tmpfile.Name())
	if err = resultspecs.Save(tmpfile, results); err != nil {
		p.Exit(xerrors.Errorf("commit "+tmpfile.Name()+": %w", err))
	}

	// write successful, move into place
	target := lockFilename(filename)
	p.FileAccess(target)
	tmpfile.Close()

	if err = os.Rename(tmpfile.Name(), target); err != nil {
		p.Exit(xerrors.Errorf("commit "+target+": %w", err))
	} else {
		p.Pass("Commit " + target)
	}
}

func compareResults(filename string, results []engine.TestResult) {
	target := lockFilename(filename)
	_, err := os.Stat(target)
	if os.IsNotExist(err) {
		p.Exit(xerrors.New("Lock file does not exists, `--commit` the first one?"))
	}

	tmpfile, err := ioutil.TempFile("", "smoke-*.yml")
	if tmpfile != nil {
		defer tmpfile.Close()
	}
	if err != nil {
		p.Exit(xerrors.Errorf("compare: %w", err))
	}

	if err = resultspecs.Save(tmpfile, results); err != nil {
		p.Exit(xerrors.Errorf("compare "+tmpfile.Name()+": %w", err))
	}

	// we defer to diff for now :p
	// TODO: Actual diffing in the CLI
	p.FileAccess("diff " + target + " " + tmpfile.Name())
	cmd := exec.Command("/usr/bin/env",
		"diff",
		"--context=3",
		target,
		tmpfile.Name())

	var (
		infile  io.WriteCloser
		outfile io.ReadCloser
		errfile io.ReadCloser
	)

	if infile, err = cmd.StdinPipe(); err != nil {
		p.Exit(xerrors.Errorf("compare "+target+": %w", err))
	} else if outfile, err = cmd.StdoutPipe(); err != nil {
		p.Exit(xerrors.Errorf("compare "+target+": %w", err))
	} else if errfile, err = cmd.StderrPipe(); err != nil {
		p.Exit(xerrors.Errorf("compare "+target+": %w", err))
	}

	go func() { _, _ = io.Copy(infile, os.Stdin) }()
	go func() { _, _ = io.Copy(os.Stdout, outfile) }()
	go func() { _, _ = io.Copy(os.Stderr, errfile) }()

	err = cmd.Run()
	if _, ok := err.(*exec.ExitError); err != nil && !ok {
		p.Exit(xerrors.Errorf("compare "+target+": %w", err))
	} else if code := cmd.ProcessState.ExitCode(); code == 0 {
		// err == nil || err is exec.ExitIfErr
		p.Pass("Stable.")
		os.Exit(0)
	} else {
		p.Fail("Changes Detected.")
		os.Exit(code) // mirror /usr/bin/diff's exit code
	}
}

func lockFilename(filename string) string {
	ext := filepath.Ext(filename)
	return filename[:len(filename)-len(ext)] + ".lock" + ext
}
