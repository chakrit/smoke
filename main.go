package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/chakrit/smoke/internal/p"
	"github.com/spf13/pflag"
)

var (
	//go:embed template.yml
	smokeTemplate []byte

	shouldShowHelp     bool
	shouldShowExpected bool

	initFile     string
	shouldList   bool
	shouldPrint  bool
	shouldCommit bool

	noColors  bool
	trackTime bool
	verbosity int
	quietness int

	includes []string
	excludes []string
)

func main() {
	pflag.BoolVarP(&shouldShowHelp, "help", "h", false, "Show help on usages.")
	pflag.BoolVarP(&shouldShowExpected, "show-expected", "s", false, "Show currently locked results without running the tests.")

	pflag.StringVar(&initFile, "init", "", "Write a starter tests.yml file; pass a path to write elsewhere.")
	pflag.Lookup("init").NoOptDefVal = "tests.yml"
	pflag.BoolVarP(&shouldList, "list", "l", false, "List all discovered tests and exit.")
	pflag.BoolVarP(&shouldPrint, "print", "p", false, "Print raw test results to stdout for scripting purposes.")
	pflag.BoolVarP(&shouldCommit, "commit", "c", false, "Commit all test output.")

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

	p.Configure(!noColors, trackTime, verbosity, quietness)

	if initFile != "" {
		if args := pflag.Args(); len(args) > 0 {
			p.Usage(fmt.Sprintf("use --init=PATH to write elsewhere; unexpected argument %q", args[0]))
			os.Exit(p.ExitUsage)
		}
		initSpec(initFile)
		return
	}

	// TODO: Make possible? Might need to overhaul test and result collection to allow
	// partial tests/modifications
	if shouldCommit && (len(includes) > 0 || len(excludes) > 0) {
		p.Usage("cannot commit partial results when using --include or --exclude")
		os.Exit(p.ExitUsage)
		return
	}

	filenames := pflag.Args()
	if len(filenames) < 1 {
		p.Usage("requires a spec filename.")
		os.Exit(p.ExitUsage)
		return
	}

	defer p.Bye()

	for _, filename := range filenames {
		processFile(filename)
	}
}

func initSpec(target string) {
	f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if os.IsExist(err) {
		p.Exit(fmt.Errorf("%s already exists", target))
	} else if err != nil {
		p.Exit(err)
	}
	defer f.Close()

	if _, err := f.Write(smokeTemplate); err != nil {
		p.Exit(err)
	}
	p.Pass("Wrote " + target)
}
