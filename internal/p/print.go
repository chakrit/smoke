package p

import (
	"fmt"
	"io"
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

		cEqual = ansi.LightBlack
		cAdded = ansi.Green
		cRemoved = ansi.Red
		cInnerChanges = ansi.LightBlack
	}
}

// output writes the drift/match report to stdout. outputErr routes operational
// diagnostics to stderr so consumers can separate "output drifted" from "SMOKE
// broke" by stream — see docs/spec/exit-codes.md.
func output(level int, s string, args ...interface{})    { outputTo(os.Stdout, level, s, args...) }
func outputErr(level int, s string, args ...interface{}) { outputTo(os.Stderr, level, s, args...) }

func outputTo(w io.Writer, level int, s string, args ...interface{}) {
	if level >= verbosity {
		return
	}

	if !startTime.IsZero() {
		dur := time.Since(startTime)
		_, _ = fmt.Fprintf(w, "%20s ", dur)
	}

	if len(args) == 0 {
		_, _ = io.WriteString(w, s)
	} else {
		_, _ = io.WriteString(w, fmt.Sprintf(s, args...))
	}
	_, _ = io.WriteString(w, "\n")
}
