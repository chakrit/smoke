package p

import (
	"fmt"
	"os"
	"strings"

	"github.com/chakrit/smoke/checks"
	"github.com/chakrit/smoke/engine"
	"github.com/chakrit/smoke/resultspecs"
)

// Exit codes — frozen contract, see docs/spec/exit-codes.md.
const (
	ExitUnchanged = 0  // output matched the lock
	ExitChanged   = 1  // drift detected
	ExitTrouble   = 2  // operational error (runner crash, I/O)
	ExitNew       = 3  // no lock; first run unreviewed
	ExitUsage     = 64 // invalid invocation (EX_USAGE)
	ExitDataErr   = 65 // spec or lock file read but malformed (EX_DATAERR)
)

// utility CLI logs
func Usage(s string)  { fmt.Fprintln(os.Stderr, s) }
func Bye()            { output(2, cLowkey+"exited."+cReset) }
func Error(err error) { outputErr(-1, cError+"ERR "+err.Error()+cReset) }
func Action(s string) { output(1, cAction+"≋≋> "+strings.ToUpper(s)+cReset) }

func Exit(err error) {
	Error(err)
	os.Exit(ExitTrouble)
}

// DataErr exits when a spec or lock file is read but won't parse or validate —
// malformed input, distinct from operational trouble. See exit-codes.md.
func DataErr(err error) {
	Error(err)
	os.Exit(ExitDataErr)
}

// testing flow
func Test(t *engine.Test)                                          { output(1, cTitle+"==> "+cTitleEm+t.Name+cReset) }
func Command(_ *engine.Test, cmd engine.Command)                   { output(2, cSubtitle+"--> "+string(cmd)+cReset) }
func Check(_ *engine.Test, _ engine.Command, chk checks.Interface) { /* no-op */ }

func TestResult(_ engine.TestResult, err error) {
	if err != nil {
		Error(err)
	}
}

func CommandResult(result engine.CommandResult, err error) {
	if err != nil {
		Error(err)
	}
}

func CheckResult(result checks.Result, err error) {
	if err != nil {
		Error(err)
		return
	}

	lines := strings.Split(string(result.Data), "\n")
	for _, line := range lines {
		output(3, fmt.Sprintf("%s  > %s:%s %s", cLowkey, result.Check.Spec(), cReset, line))
	}
}

// lockfile flow
func FileAccess(f string) { output(2, cSubtitle+"--> "+f+cReset) }
func Pass(s string)       { output(-1, cPass+"\n  ✔ "+s+"\n"+cReset) }

// drift verdicts — neutral vocabulary, no pass/fail framing. UNCHANGED means
// drift-free, not verified-correct; CHANGED means review, not failure.
func Unchanged(lock string) {
	output(-1, "\nUNCHANGED — output matches "+lock+" (drift-free, not verified correct)\n")
}
func Changed(lock string) {
	output(-1, cAction+"\nCHANGED — output drifted from "+lock+"; review and re-commit if intended\n"+cReset)
}
func New(lock string) {
	output(-1, cAction+"\nNEW — no lock at "+lock+"; first run unreviewed, --commit to create the golden\n"+cReset)
}

// diff flow
func TestEdit(edit resultspecs.TestEdit) {
	c, prefix := colorByAction(edit.Action)
	output(0, c+prefix+" ==> "+edit.Name+cReset)
}

func CommandEdit(edit resultspecs.CommandEdit) {
	c, prefix := colorByAction(edit.Action)
	output(0, c+prefix+" --> "+edit.Name+cReset)
}

func CheckEdit(edit resultspecs.CheckEdit) {
	c, prefix := colorByAction(edit.Action)
	output(0, c+prefix+"   > "+edit.Name+cReset)
}

func LineEdit(edit resultspecs.LineEdit) {
	c, prefix := colorByAction(edit.Action)
	output(0, c+prefix+"       "+edit.Line+cReset)
}

func colorByAction(action resultspecs.Action) (string, string) {
	switch action {
	case resultspecs.NoOp:
		return cEqual, "   "
	case resultspecs.Equal:
		return cEqual, "   "
	case resultspecs.Added:
		return cAdded, "+++"
	case resultspecs.Removed:
		return cRemoved, "---"
	case resultspecs.InnerChanges:
		return cInnerChanges, "   "
	default:
		panic("bad edit action: " + fmt.Sprint(action))
	}
}
