package p

import (
	"fmt"
	"os"
	"strings"

	"github.com/chakrit/smoke/engine"
	"github.com/chakrit/smoke/resultspecs"
	"github.com/mgutz/ansi"
)

// utility CLI logs
func Usage(s string)  { fmt.Fprintln(os.Stderr, s) }
func Bye()            { output(1, cLowkey+"exited."+cReset) }
func Error(err error) { output(-1, cError+"ERR "+err.Error()+cReset) }
func Action(s string) { output(1, cAction+"≋≋> "+strings.ToUpper(s)+cReset) }

func Exit(err error) {
	Error(err)
	os.Exit(1)
}

// testing flow
func Test(t *engine.Test)                        { output(1, cTitle+"==> "+cTitleEm+t.Name+cReset) }
func Command(_ *engine.Test, cmd engine.Command) { output(2, cSubtitle+"--> "+string(cmd)+cReset) }

func TestResult(_ engine.TestResult, err error) {
	if err != nil {
		Error(err)
	}
}

func CommandResult(result engine.CommandResult, err error) {
	if err != nil {
		Error(err)
		return
	}

	for _, chk := range result.Checks {
		lines := strings.Split(string(chk.Data), "\n")
		for _, line := range lines {
			output(3, ansi.LightBlack+"%8s:"+ansi.Reset+" %s",
				chk.Check.Name(), line)
		}
	}
}

// lockfile flow
func FileAccess(f string) { output(2, cSubtitle+"--> "+f+cReset) }
func Pass(s string)       { output(-1, cPass+"\n  ✔ "+s+"\n"+cReset) }
func Fail(s string)       { output(-1, cFail+"\n  ✘ "+s+"\n"+cReset) }

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
	output(0, c+prefix+"     "+edit.Name+cReset)
}

func LineEdit(edit resultspecs.LineEdit) {
	c, prefix := colorByAction(edit.Action)
	output(0, c+prefix+"       "+edit.Line+cReset)
}

func colorByAction(action resultspecs.Action) (string, string) {
	switch action {
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
