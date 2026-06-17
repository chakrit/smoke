package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chakrit/smoke/engine"
	"github.com/chakrit/smoke/internal"
	"github.com/chakrit/smoke/internal/p"
	"github.com/chakrit/smoke/resultspecs"
	"github.com/chakrit/smoke/testspecs"
)

func processFile(filename string) {
	if shouldShowExpected {
		showResults(filename)
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		p.Exit(fmt.Errorf(filename+": %w", err))
	}
	defer file.Close()

	tests, err := testspecs.Load(file, filename)
	if err != nil {
		p.DataErr(fmt.Errorf(filename+": %w", err))
	}

	if len(includes) > 0 {
		tests = internal.Whitelist(tests, includes, func(t *engine.Test) string {
			return t.Name
		})
	}
	if len(excludes) > 0 {
		tests = internal.Blacklist(tests, excludes, func(t *engine.Test) string {
			return t.Name
		})
	}

	if shouldList {
		listTests(tests)
		return
	}

	p.Action("Running Tests")
	results := runTests(tests)
	if shouldPrint {
		p.Action("Printing Result")
		printResults(results)
	} else if shouldCommit {
		p.Action("Writing Lock File")
		commitResults(filename, results)
	} else {
		p.Action("Comparing Lock File")
		compareResults(filename, results)
	}
}

func listTests(tests []*engine.Test) {
	v := p.Verbosity()

	for _, test := range tests {
		fmt.Fprintln(os.Stdout, test.Name)
		if v > 1 {
			for _, cmd := range test.Commands {
				fmt.Fprintf(os.Stdout, "\t%s\n", cmd)
			}
		}
	}
}

func showResults(filename string) {
	target := lockFilename(filename)
	p.FileAccess(target)

	file, err := os.Open(target)
	if err != nil {
		p.Exit(err)
		return
	}
	defer file.Close()

	results, err := resultspecs.Load(file)
	if err != nil {
		p.DataErr(fmt.Errorf(target+": %w", err))
	}

	if len(includes) > 0 {
		results = internal.Whitelist(results, includes, func(r resultspecs.TestResultSpec) string {
			return r.Name
		})
	}
	if len(excludes) > 0 {
		results = internal.Blacklist(results, excludes, func(r resultspecs.TestResultSpec) string {
			return r.Name
		})
	}

	for _, test := range results {
		p.TestEdit(resultspecs.TestEdit{Name: test.Name, Action: resultspecs.NoOp})
		for _, cmd := range test.Commands {
			p.CommandEdit(resultspecs.CommandEdit{Name: cmd.Command, Action: resultspecs.NoOp})
			for _, check := range cmd.Checks {
				p.CheckEdit(resultspecs.CheckEdit{Name: check.Name, Action: resultspecs.NoOp})
				for _, line := range check.Data {
					p.LineEdit(resultspecs.LineEdit{Line: line, Action: resultspecs.NoOp})
				}
			}
		}
	}
}

func runTests(tests []*engine.Test) []engine.TestResult {
	var (
		failed = false
		hooks  = Hooks{
			WrapErr: func(t *engine.Test, err error) error {
				return fmt.Errorf(t.Name+": %w", err)
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
		os.Exit(p.ExitTrouble)
		return nil
	}

	return results
}

func printResults(results []engine.TestResult) {
	if err := resultspecs.Save(os.Stdout, results); err != nil {
		p.Exit(fmt.Errorf("print to stdout: %w", err))
	}
	p.Pass("Print")
}

func commitResults(filename string, results []engine.TestResult) {
	tmpfile, err := os.CreateTemp("", "smoke-*.yml")
	if err != nil {
		p.Exit(fmt.Errorf("commit: %w", err))
	}
	defer tmpfile.Close()

	p.FileAccess(tmpfile.Name())
	if err = resultspecs.Save(tmpfile, results); err != nil {
		p.Exit(fmt.Errorf("commit "+tmpfile.Name()+": %w", err))
	}

	// write successful, move into place
	target := lockFilename(filename)
	p.FileAccess(target)
	tmpfile.Close()

	if err = os.Rename(tmpfile.Name(), target); err != nil {
		p.Exit(fmt.Errorf("commit "+target+": %w", err))
	} else {
		p.Pass("Commit " + target)
	}
}

func compareResults(filename string, results []engine.TestResult) {
	lockname := lockFilename(filename)

	var out reporter = consoleReporter{}
	if shouldJSON {
		out = jsonReporter{w: os.Stdout}
	}
	report := func(st status, edits []resultspecs.TestEdit) {
		if err := out.Report(lockname, st, edits); err != nil {
			p.Exit(fmt.Errorf("report: %w", err))
		}
		os.Exit(st.ExitCode())
	}

	lockfile, err := os.Open(lockname)
	if os.IsNotExist(err) {
		report(statusNew, nil)
	} else if err != nil {
		p.Exit(fmt.Errorf(lockname+": %w", err))
	} else {
		defer lockfile.Close()
	}

	lockspec, err := resultspecs.Load(lockfile)
	if err != nil {
		p.DataErr(fmt.Errorf(lockname+": %w", err))
	}

	// if includes/excludes are set, only compare those, otherwise the excluded/included
	// tests are also compared even though they havn't been ran
	if len(includes) > 0 {
		lockspec = internal.Whitelist(lockspec, includes, func(s resultspecs.TestResultSpec) string {
			return s.Name
		})
	}
	if len(excludes) > 0 {
		lockspec = internal.Blacklist(lockspec, excludes, func(s resultspecs.TestResultSpec) string {
			return s.Name
		})
	}

	var newspecs []resultspecs.TestResultSpec
	for _, result := range results {
		if spec, err := resultspecs.FromTestResult(result); err != nil {
			p.Exit(fmt.Errorf("compare: %w", err))
		} else {
			newspecs = append(newspecs, spec)
		}
	}

	edits, differs, err := resultspecs.Compare(lockspec, newspecs)
	if err != nil {
		p.Exit(fmt.Errorf("compare: %w", err))
	}

	st := statusUnchanged
	if differs {
		st = statusChanged
	}
	report(st, edits)
}

func lockFilename(filename string) string {
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]
	// results are always serialized as YAML, so only YAML specs keep their
	// extension; every other format (.cue, .json, .jsonl) locks to .lock.yml
	// rather than a .lock.<ext> we could never round-trip.
	switch ext {
	case ".yml", ".yaml":
		return base + ".lock" + ext
	default:
		return base + ".lock.yml"
	}
}
