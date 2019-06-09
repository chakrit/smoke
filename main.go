package main

import (
	"os"

	"github.com/chakrit/smoke/engine"
	"github.com/chakrit/smoke/internal/p"
	"github.com/chakrit/smoke/specs"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var (
	lockFile       string
	shouldShowHelp bool
	shouldList     bool
	shouldCommit   bool
)

func main() {
	pflag.BoolVarP(&shouldShowHelp, "help", "h", false, "Show help on usages.")
	pflag.BoolVarP(&shouldList, "list", "l", false, "List all discovered tests and exit.")
	pflag.BoolVarP(&shouldCommit, "commit", "c", false, "Commit all checked test output.")
	pflag.StringVarP(&lockFile, "lockfile", "f", "", "Filename to read lock result from (or write to, when committing).")
	pflag.Parse()

	if shouldShowHelp {
		pflag.Usage()
		return
	}

	filenames := pflag.Args()
	if len(filenames) < 1 {
		p.UsageHint("requires a spec filename.")
		os.Exit(1)
		return
	}

	defer p.Bye()
	for _, filename := range filenames {
		processFile(filename)
	}
}

func processFile(filename string) {
	file, err := specs.Load(filename)
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
		for _, test := range tests {
			p.TestDesc(filename, test)
		}
		return
	}

	var results []engine.TestResult
	for _, test := range tests {
		p.BeforeTest(filename, test)
		if result, err := engine.RunTest(test); err != nil {
			p.Error(err)
		} else {
			p.AfterTest(filename, test, result)
			results = append(results, result)
		}
	}
}
