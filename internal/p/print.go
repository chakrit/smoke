package p

import (
	"fmt"
	"os"
	"strings"

	"github.com/chakrit/smoke/engine"
)

var trim = strings.TrimSpace

func UsageHint(s string) {
	fmt.Fprintln(os.Stderr, s)
}

func Bye() {
	fmt.Println("exited.")
}

func Error(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func TestDesc(filename string, t *engine.Test) {
	fmt.Printf("%s: %s\n", filename, t.Name)
	for _, cmd := range t.Commands {
		fmt.Printf("\t%s\n", cmd)
	}
}

func BeforeTest(filename string, t *engine.Test) {
	fmt.Printf("%s: %s\n", filename, t.Name)
}

func AfterTest(filename string, t *engine.Test, result engine.TestResult) {
	for _, cmd := range result.Commands {
		cmdstr := trim(string(cmd.Command))
		if cmd.Err != nil {
			fmt.Printf("\t%s: %s\n", cmdstr, cmd.Err)
			continue
		}

		fmt.Printf("\t%s:\n", cmdstr)
		for _, chk := range cmd.Checks {
			fmt.Printf("\t\t%s:\t%s\n", chk.Name, trim(string(chk.Data)))
		}
	}
}
