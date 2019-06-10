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
	shouldCommit   bool
	noColors       bool
)

func main() {
	pflag.BoolVarP(&shouldShowHelp, "help", "h", false, "Show help on usages.")
	pflag.BoolVarP(&shouldList, "list", "l", false, "List all discovered tests and exit.")
	pflag.BoolVarP(&shouldCommit, "commit", "c", false, "Commit all test output.")
	pflag.StringVarP(&lockFile, "lockfile", "f", "", "Filename to read lock result from (or write to, when committing).")
	pflag.BoolVar(&noColors, "no-color", false, "Turns off console coloring.")
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

	p.DisableColors(noColors)
	defer p.Bye()

	for _, filename := range filenames {
		processFile(filename)
	}
}
