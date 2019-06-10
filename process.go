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
	"github.com/pkg/errors"
)

func processFile(filename string) {
	file, err := os.Open(filename)
	p.ExitError(errors.Wrap(err, filename))

	defer file.Close()

	tests, err := testspecs.Load(file, filename)
	p.ExitError(errors.Wrap(err, filename))

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
		run     engine.Runner = engine.DefaultRunner{Hooks: p.Hooks{}}
		results []engine.TestResult
		failed  = false
	)

	for _, test := range tests {
		if result, err := run.Test(test); err != nil {
			// error already printed by p.Hooks{}
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
	p.ExitError(errors.Wrap(err, "commit"))

	defer tmpfile.Close()

	p.FileAccess(tmpfile.Name())
	err = resultspecs.Save(tmpfile, results)
	p.ExitError(errors.Wrap(err, "commit"))

	// write successful, move into place
	target := lockFilename(filename)
	p.FileAccess(target)
	tmpfile.Close()

	err = os.Rename(tmpfile.Name(), target)
	p.ExitError(errors.Wrap(err, "commit"))

	p.Pass("Commit " + target)
}

func compareResults(filename string, results []engine.TestResult) {
	target := lockFilename(filename)
	_, err := os.Stat(target)
	if os.IsNotExist(err) {
		p.ExitError(errors.New("Lock file does not exists, `--commit` the first one?"))
	}

	tmpfile, err := ioutil.TempFile("", "smoke-*.yml")
	p.ExitError(errors.Wrap(err, "compare"))
	defer tmpfile.Close()

	err = resultspecs.Save(tmpfile, results)
	p.ExitError(errors.Wrap(err, "compare"))

	// we defer to diff for now :p
	// TODO: Actual diffing in the CLI
	p.FileAccess("diff " + target + " " + tmpfile.Name())
	cmd := exec.Command("diff", target, tmpfile.Name())

	inpipe, err := cmd.StdinPipe()
	p.ExitError(errors.Wrap(err, "compare"))
	outfile, err := cmd.StdoutPipe()
	p.ExitError(errors.Wrap(err, "compare"))
	errfile, err := cmd.StderrPipe()
	p.ExitError(errors.Wrap(err, "compare"))

	go func() { _, _ = io.Copy(inpipe, os.Stdin) }()
	go func() { _, _ = io.Copy(os.Stdout, outfile) }()
	go func() { _, _ = io.Copy(os.Stderr, errfile) }()

	err = cmd.Run()
	if _, ok := err.(*exec.ExitError); err != nil && !ok {
		p.ExitError(errors.Wrap(err, "compare"))
	} else if code := cmd.ProcessState.ExitCode(); code == 0 {
		p.Pass("Stable.")
		os.Exit(0)
	} else {
		p.Fail("Changes Detected.")
		os.Exit(code)
	}
}

func lockFilename(filename string) string {
	ext := filepath.Ext(filename)
	return filename[:len(filename)-len(ext)] + ".lock" + ext
}
