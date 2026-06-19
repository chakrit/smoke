package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chakrit/smoke/engine"
	"github.com/chakrit/smoke/internal"
	"github.com/chakrit/smoke/internal/p"
	"github.com/chakrit/smoke/resultspecs"
	"github.com/chakrit/smoke/runcache"
	"github.com/chakrit/smoke/testspecs"
)

// processFile runs one spec and returns its compare verdict plus any fatal.
// main is the single exit authority: it fail-fasts on a fatal (dataError → 65,
// any other error → 2) and otherwise folds the verdict into the aggregate exit.
// Non-compare modes (list/print/commit/show) report at the site and return
// statusUnchanged on success. Diagnostics surface here or live via the run
// hooks; main maps the error to its code without re-printing.
func processFile(filename string) (status, error) {
	if shouldShowExpected {
		return statusUnchanged, showResults(filename)
	}
	if shouldCommitLast {
		return statusUnchanged, commitLast(filename)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return statusUnchanged, fmt.Errorf(filename+": %w", err)
	}

	tests, err := testspecs.Load(bytes.NewReader(content), filename)
	if err != nil {
		return statusUnchanged, dataErr(fmt.Errorf(filename+": %w", err))
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
		return statusUnchanged, nil
	}

	p.Action("Running Tests")
	results, err := runTests(tests)
	if err != nil {
		return statusUnchanged, err
	}
	saveRunCache(filename, content, results)

	switch {
	case shouldPrint:
		p.Action("Printing Result")
		return statusUnchanged, printResults(results)
	case shouldCommit:
		p.Action("Writing Lock File")
		return statusUnchanged, commitResults(filename, results)
	default:
		p.Action("Comparing Lock File")
		return compareResults(filename, results)
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

func showResults(filename string) error {
	target := lockFilename(filename)
	p.FileAccess(target)

	file, err := os.Open(target)
	if err != nil {
		return fmt.Errorf(target+": %w", err)
	}
	defer file.Close()

	results, err := resultspecs.Load(file)
	if err != nil {
		return dataErr(fmt.Errorf(target+": %w", err))
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
	return nil
}

// runTests runs each test in order, stopping at the first failure. The run hooks
// already surface the error live, so it comes back wrapped in reported{} — main
// fail-fasts (exit 2) without printing it again.
func runTests(tests []*engine.Test) ([]engine.TestResult, error) {
	hooks := Hooks{
		WrapErr: func(t *engine.Test, err error) error {
			return fmt.Errorf(t.Name+": %w", err)
		},
	}
	run := engine.DefaultRunner{Hooks: hooks}

	var results []engine.TestResult
	for _, test := range tests {
		result, err := run.Test(test)
		if err != nil {
			return nil, reportedErr(err)
		}
		results = append(results, result)
	}
	return results, nil
}

func printResults(results []engine.TestResult) error {
	specs, err := resultspecs.FromTestResults(results)
	if err != nil {
		return fmt.Errorf("print to stdout: %w", err)
	}
	if err := resultspecs.Save(os.Stdout, specs); err != nil {
		return fmt.Errorf("print to stdout: %w", err)
	}
	p.Pass("Print")
	return nil
}

func commitResults(filename string, results []engine.TestResult) error {
	target := lockFilename(filename)

	specs, err := resultspecs.FromTestResults(results)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	// A filtered run observed only a subset, so merge it onto the existing lock
	// to keep the golden for tests it never ran. An unfiltered run is the whole
	// truth and replaces the lock wholesale, pruning tests dropped from the spec.
	if filtering() {
		base, err := loadLockSpecs(target)
		if err != nil {
			return err
		}
		specs = resultspecs.Merge(base, specs)
	}

	return writeLock(target, specs)
}

// commitLast commits the previous run from cache without re-running, after
// verifying the cached run still matches the current spec — SMOKE never blesses
// a golden for a spec that has changed since it was observed.
func commitLast(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf(filename+": %w", err)
	}

	snap, ok, err := runcache.Load(filename)
	if err != nil {
		return fmt.Errorf("commit-last: %w", err)
	}
	if !ok {
		return dataErr(fmt.Errorf("commit-last: no cached run for %s; run it once first", filename))
	}
	if snap.SpecHash != runcache.HashSpec(content) {
		return dataErr(fmt.Errorf("commit-last: %s changed since the last run; re-run before committing", filename))
	}

	target := lockFilename(filename)
	specs := snap.Results
	if snap.Partial {
		base, err := loadLockSpecs(target)
		if err != nil {
			return err
		}
		specs = resultspecs.Merge(base, specs)
	}
	return writeLock(target, specs)
}

// loadLockSpecs reads an existing lock, returning nil when none exists yet (a
// first partial commit has no prior golden to preserve).
func loadLockSpecs(target string) ([]resultspecs.TestResultSpec, error) {
	file, err := os.Open(target)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(target+": %w", err)
	}
	defer file.Close()

	specs, err := resultspecs.Load(file)
	if err != nil {
		return nil, dataErr(fmt.Errorf(target+": %w", err))
	}
	return specs, nil
}

// writeLock atomically replaces the lock: write a temp file, then rename, so a
// crash mid-write never corrupts an existing lock.
func writeLock(target string, specs []resultspecs.TestResultSpec) error {
	tmpfile, err := os.CreateTemp("", "smoke-*.yml")
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	defer tmpfile.Close()

	p.FileAccess(tmpfile.Name())
	if err := resultspecs.Save(tmpfile, specs); err != nil {
		return fmt.Errorf("commit "+tmpfile.Name()+": %w", err)
	}

	p.FileAccess(target)
	tmpfile.Close()

	if err := os.Rename(tmpfile.Name(), target); err != nil {
		return fmt.Errorf("commit "+target+": %w", err)
	}
	p.Pass("Commit " + target)
	return nil
}

// saveRunCache persists the run so a later --commit-last can bless it without
// re-running. Best-effort: the cache is a convenience, so a failure here must
// never fail the run that produced it.
func saveRunCache(filename string, content []byte, results []engine.TestResult) {
	specs, err := resultspecs.FromTestResults(results)
	if err != nil {
		return
	}
	_ = runcache.Save(filename, runcache.Snapshot{
		SpecHash: runcache.HashSpec(content),
		Partial:  filtering(),
		Results:  specs,
	})
}

func compareResults(filename string, results []engine.TestResult) (status, error) {
	lockname := lockFilename(filename)

	var out reporter = consoleReporter{}
	if shouldJSON {
		out = jsonReporter{w: os.Stdout}
	}
	report := func(st status, edits []resultspecs.TestEdit) (status, error) {
		if err := out.Report(lockname, st, edits); err != nil {
			return st, fmt.Errorf("report: %w", err)
		}
		return st, nil
	}

	lockfile, err := os.Open(lockname)
	if os.IsNotExist(err) {
		return report(statusNew, nil)
	} else if err != nil {
		return statusUnchanged, fmt.Errorf(lockname+": %w", err)
	}
	defer lockfile.Close()

	lockspec, err := resultspecs.Load(lockfile)
	if err != nil {
		return statusUnchanged, dataErr(fmt.Errorf(lockname+": %w", err))
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

	newspecs, err := resultspecs.FromTestResults(results)
	if err != nil {
		return statusUnchanged, fmt.Errorf("compare: %w", err)
	}

	edits, differs, err := resultspecs.Compare(lockspec, newspecs)
	if err != nil {
		return statusUnchanged, fmt.Errorf("compare: %w", err)
	}

	st := statusUnchanged
	if differs {
		st = statusChanged
	}
	return report(st, edits)
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
