package main

import (
	"os"

	"github.com/chakrit/smoke/internal/p"
	"github.com/spf13/pflag"
)

var (
	lockFile       string
	shouldShowHelp bool
	shouldList     bool
	shouldPrint    bool
	shouldCommit   bool

	noColors  bool
	trackTime bool
	verbosity int
	quietness int

	includes []string
	excludes []string
)

func main() {
	pflag.BoolVarP(&shouldShowHelp, "help", "h", false, "Show help on usages.")

	pflag.BoolVarP(&shouldList, "list", "l", false, "List all discovered tests and exit.")
	pflag.BoolVarP(&shouldPrint, "print", "p", false, "Print results but don't do any comparison.")
	pflag.BoolVarP(&shouldCommit, "commit", "c", false, "Commit all test output.")
	pflag.StringVarP(&lockFile, "lockfile", "f", "", "Filename to read lock result from (or write to, when committing).")

	pflag.BoolVar(&noColors, "no-color", false, "Turns off console coloring.")
	pflag.BoolVar(&trackTime, "time", false, "Log timestamps.")
	pflag.CountVarP(&verbosity, "verbose", "v", "Increase log output chattiness.")
	pflag.CountVarP(&quietness, "quiet", "q", "Decrease log output chattiness.")

	pflag.StringSliceVarP(&includes, "include", "i", nil, "Only run tests matching the given pattern")
	pflag.StringSliceVarP(&excludes, "exclude", "x", nil, "Do not run tests matching the given pattern")
	pflag.Parse()

	if shouldShowHelp {
		pflag.Usage()
		return
	}

	// TODO: Make possible? Might need to overhaul test and result collection to allow
	// partial tests/modifications
	if shouldCommit && (len(includes) > 0 || len(excludes) > 0) {
		p.Usage("cannot commit partial results when using --include or --exclude")
		os.Exit(1)
		return
	}

	filenames := pflag.Args()
	if len(filenames) < 1 {
		p.Usage("requires a spec filename.")
		os.Exit(1)
		return
	}

	p.Configure(!noColors, trackTime, verbosity, quietness)
	defer p.Bye()

	for _, filename := range filenames {
		processFile(filename)
	}
}
