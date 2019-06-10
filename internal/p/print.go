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
	cDeEm     string
	cError    string
	cReset    string
)

func init() { setColors() }

func DisableColors(disable bool) {
	ansi.DisableColors(disable)
	setColors()
}

func setColors() {
	cTitle = ansi.Magenta
	cTitleEm = ansi.ColorCode("magenta+b")
	cSubtitle = ansi.Blue
	cDeEm = ansi.LightBlack
	cError = ansi.Red
	cReset = ansi.Reset
}

// utility CLI logs
func UsageHint(s string) {
	fmt.Fprintln(os.Stderr, s)
}

func Bye() {
	fmt.Println(cDeEm+"exited.", cReset)
}

func Error(err error) {
	fmt.Fprintln(os.Stderr, cError+err.Error(), cReset)
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
