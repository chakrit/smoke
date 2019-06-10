package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/chakrit/smoke/engine"
	"github.com/chakrit/smoke/internal/p"
	"github.com/chakrit/smoke/testspecs"
	"github.com/pkg/errors"
)

func processFile(filename string) {
	file, err := testspecs.Load(filename)
	if err != nil {
		p.Error(errors.Wrap(err, filename))
		return
	}

	tests, err := file.Tests()
	if err != nil {
		p.Error(errors.Wrap(err, filename))
		return
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
	} // else { diffResults(results) }
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
	if err != nil {
		p.Error(errors.Wrap(err, "commit"))
		os.Exit(1)
		return
	}

	defer tmpfile.Close()

	p.WriteFile(tmpfile.Name())
	if err := engine.SaveAll(tmpfile, results); err != nil {
		p.Error(errors.Wrap(err, "commit"))
		os.Exit(1)
		return
	}

	// write succesful, move into place
	target := lockFilename(filename)
	p.WriteFile(target)
	tmpfile.Close()

	if err := os.Rename(tmpfile.Name(), target); err != nil {
		p.Error(errors.Wrap(err, "commit"))
		os.Exit(1)
		return
	}

}

func lockFilename(filename string) string {
	ext := filepath.Ext(filename)
	return filename[:len(ext)+1] + ".lock" + ext
}
