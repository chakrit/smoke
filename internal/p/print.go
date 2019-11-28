package p

import (
	"fmt"
	"os"
	"time"

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

	startTime time.Time
	verbosity int
)

func init() { Configure(true, false, 1, 0) }

func Verbosity() int { return verbosity }

func Configure(color, trackTime bool, v int, q int) {
	verbosity = 1 + v - q
	if trackTime {
		startTime = time.Now()
	}

	if !color {
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

	if !startTime.IsZero() {
		dur := time.Now().Sub(startTime)
		_, _ = fmt.Fprintf(os.Stdout, "%20s ", dur)
	}

	if len(args) == 0 {
		_, _ = os.Stdout.WriteString(s)
	} else {
		_, _ = os.Stdout.WriteString(fmt.Sprintf(s, args...))
	}
	_, _ = os.Stdout.WriteString("\n")
}
