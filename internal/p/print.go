package p

import (
	"fmt"
	"os"

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

	cEqual        string
	cAdded        string
	cRemoved      string
	cInnerChanges string

	verbosity int
)

func init() { Configure(true, 1, 0) }

func Configure(colored bool, v int, q int) {
	verbosity = 1 + v - q

	if !colored {
		cTitle = ""
		cTitleEm = ""
		cSubtitle = ""
		cLowkey = ""
		cReset = ""

		cAction = ""
		cError = ""
		cPass = ""
		cFail = ""

		cEqual = ""
		cAdded = ""
		cRemoved = ""
		cInnerChanges = ""
	} else {
		cTitle = ansi.Magenta
		cTitleEm = ansi.ColorCode("magenta+b")
		cSubtitle = ansi.Blue
		cLowkey = ansi.LightBlack
		cReset = ansi.Reset

		cAction = ansi.ColorCode("cyan+b")
		cError = ansi.Red
		cPass = ansi.ColorCode("green+b")
		cFail = ansi.ColorCode("red+b")

		cEqual = ansi.LightBlack
		cAdded = ansi.Green
		cRemoved = ansi.Red
		cInnerChanges = ansi.LightBlack
	}
}

func output(level int, s string, args ...interface{}) {
	if level >= verbosity {
		return
	}

	if len(args) == 0 {
		_, _ = os.Stdout.WriteString(s)
	} else {
		_, _ = os.Stdout.WriteString(fmt.Sprintf(s, args...))
	}
	_, _ = os.Stdout.WriteString("\n")
}
