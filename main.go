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

	initFile         string
	shouldList       bool
	shouldPrint      bool
	shouldCommit     bool
	shouldCommitLast bool
	shouldJSON       bool

	noColors  bool
	trackTime bool
	verbosity int
	quietness int

	includes []string
	excludes []string
)

// usageHeader frames SMOKE as a drift detector, not a pass/fail test runner —
// the same framing the exit-code contract encodes. See docs/spec/exit-codes.md.
const usageHeader = `SMOKE is a drift detector for command output, not an assertion engine.
A clean run means output matched the committed golden — UNCHANGED is not "correct".

Usage:
  smoke [flags] <spec.yml>...

Flags:
`

const usageExitCodes = `
Exit codes:
  0   UNCHANGED   output matched the lock (not "tests passed")
  1   CHANGED     drift — output moved from the golden (includes MISSING, timeout)
  2               operational error — SMOKE itself broke (runner crash, I/O)
  3   NEW         no lock yet; first run is unreviewed
  64              usage error — invalid invocation (bad flags)
  65              data error — spec or lock file is malformed
`

func usage() {
	fmt.Fprint(os.Stderr, usageHeader)
	pflag.PrintDefaults()
	fmt.Fprint(os.Stderr, usageExitCodes)
}

func main() {
	pflag.Usage = usage
	// ContinueOnError makes pflag return parse errors instead of exiting 2 itself
	// (its ExitOnError default), so a bad flag routes through our ExitUsage (64).
	pflag.CommandLine.Init(os.Args[0], pflag.ContinueOnError)

	pflag.BoolVarP(&shouldShowHelp, "help", "h", false, "Show help on usages.")
	pflag.BoolVarP(&shouldShowExpected, "show-expected", "s", false, "Show currently locked results without running the tests.")

	pflag.StringVar(&initFile, "init", "", "Write a starter tests.yml file; pass a path to write elsewhere.")
	pflag.Lookup("init").NoOptDefVal = "tests.yml"
	pflag.BoolVarP(&shouldList, "list", "l", false, "List all discovered tests and exit.")
	pflag.BoolVarP(&shouldPrint, "print", "p", false, "Print raw test results to stdout for scripting purposes.")
	pflag.BoolVarP(&shouldCommit, "commit", "c", false, "Commit all test output.")
	pflag.BoolVar(&shouldCommitLast, "commit-last", false, "Commit the previous run's results without re-running; refuses if the spec changed.")
	pflag.BoolVar(&shouldJSON, "json", false, "Emit the compare result as a machine-readable JSON document.")

	pflag.BoolVar(&noColors, "no-color", false, "Turns off console coloring.")
	pflag.BoolVar(&trackTime, "time", false, "Log timestamps.")
	pflag.CountVarP(&verbosity, "verbose", "v", "Increase log output chattiness.")
	pflag.CountVarP(&quietness, "quiet", "q", "Decrease log output chattiness.")

	pflag.StringSliceVarP(&includes, "include", "i", nil, "Only run tests matching the given pattern")
	pflag.StringSliceVarP(&excludes, "exclude", "x", nil, "Do not run tests matching the given pattern")
	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		p.Usage(err.Error())
		pflag.Usage()
		os.Exit(p.ExitUsage)
	}

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

	// --json renders the compare outcome only; mixing it with another output mode
	// has no defined meaning, so reject rather than silently ignore it.
	if shouldJSON && (shouldList || shouldPrint || shouldCommit || shouldShowExpected) {
		p.Usage("--json applies to the default compare mode; cannot combine with --list/--print/--commit/--show-expected")
		os.Exit(p.ExitUsage)
		return
	}

	// --commit-last is its own mode: it replays the previous run rather than
	// running, so it owns neither another output mode nor a fresh filter.
	if shouldCommitLast && (shouldCommit || shouldPrint || shouldList || shouldShowExpected || shouldJSON) {
		p.Usage("--commit-last commits the previous run; cannot combine with another mode")
		os.Exit(p.ExitUsage)
		return
	}
	if shouldCommitLast && filtering() {
		p.Usage("--commit-last replays the previous run's scope; cannot combine with --include/--exclude")
		os.Exit(p.ExitUsage)
		return
	}

	filenames := pflag.Args()
	if len(filenames) < 1 {
		p.Usage("requires a spec filename.")
		os.Exit(p.ExitUsage)
		return
	}

	// Single exit authority: every spec is processed (drift in spec N never hides
	// behind spec N-1), each verdict folds into one exit (any drift → non-zero),
	// and a fatal (malformed spec → 65, operational → 2) fail-fasts the run
	// because each spec may carry side effects the next depends on. See
	// docs/spec/exit-codes.md.
	verdict := statusUnchanged
	for _, filename := range filenames {
		st, err := processFile(filename)
		if err != nil {
			if !wasReported(err) {
				p.Error(err)
			}
			os.Exit(exitCode(err))
		}
		verdict = verdict.Merge(st)
	}

	p.Bye()
	os.Exit(verdict.ExitCode())
}

// filtering reports whether the run is scoped to a subset of tests, which makes
// a commit a partial merge rather than a wholesale overwrite.
func filtering() bool {
	return len(includes) > 0 || len(excludes) > 0
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
