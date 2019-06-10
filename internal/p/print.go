package p

import (
	"fmt"
	"os"
	"strings"

	"github.com/chakrit/smoke/engine"
	"github.com/mgutz/ansi"
)

var ( // stylesheet :p
	cTitle    string
	cTitleEm  string
	cSubtitle string
	cLowkey   string
	cReset    string

	cAction string
	cError  string
	cPass   string
	cFail   string
)

func init() { setColors(false) }

func DisableColors(disable bool) {
	setColors(disable)
}

func setColors(disable bool) {
	if disable {
		cAction = ""
		cTitle = ""
		cTitleEm = ""
		cSubtitle = ""
		cLowkey = ""
		cError = ""
		cReset = ""
		cPass = ""
		cFail = ""
	} else {
		cAction = ansi.ColorCode("cyan+b")
		cTitle = ansi.Magenta
		cTitleEm = ansi.ColorCode("magenta+b")
		cSubtitle = ansi.Blue
		cLowkey = ansi.LightBlack
		cError = ansi.Red
		cReset = ansi.Reset
		cPass = ansi.ColorCode("green+b")
		cFail = ansi.ColorCode("red+b")
	}
}

// utility CLI logs
func UsageHint(s string) {
	fmt.Fprintln(os.Stderr, s)
}

func Bye() {
	fmt.Println(cLowkey+"exited.", cReset)
}

func ExitError(err error) {
	if err == nil {
		return
	}

	fmt.Fprintln(os.Stderr, cError+"ERR", err.Error(), cReset)
	os.Exit(1)
}

func Action(s string) {
	fmt.Println(cAction+"≋≋>", strings.ToUpper(s), cReset)
}

// testing flow
func Test(t *engine.Test) {
	fmt.Println(cTitle+"==>", cTitleEm+t.Name, cReset)
}

func Command(_ *engine.Test, cmd engine.Command) {
	fmt.Println(cSubtitle+"-->", cmd, cReset)
}

func TestResult(_ engine.TestResult, err error) {
	if err != nil {
		fmt.Println(ansi.Red, "ERR", err, ansi.Reset)
		return
	}
}

func CommandResult(result engine.CommandResult, err error) {
	if err != nil {
		fmt.Println("ERR", err)
		return
	}

	for _, chk := range result.Checks {
		lines := strings.Split(string(chk.Data), "\n")
		for _, line := range lines {
			fmt.Printf(ansi.LightBlack+"%8s:"+ansi.Reset+" %s\n",
				chk.Name, line)
		}
	}
}

// lockfile flow
func FileAccess(filename string) {
	fmt.Println(cSubtitle+"-->", filename, cReset)
}

func Pass(s string) {
	fmt.Println(cPass+"  ✔", s, cReset)
}

func Fail(s string) {
	fmt.Println(cFail+"  ✘", s, cReset)
}
