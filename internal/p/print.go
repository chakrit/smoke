package p

import (
	"fmt"
	"os"
	"strings"

	"github.com/chakrit/smoke/engine"
)

var trim = strings.TrimSpace

// utility CLI logs
func UsageHint(s string) {
	fmt.Fprintln(os.Stderr, s)
}

func Bye() {
	fmt.Println("exited.")
}

func Error(err error) {
	fmt.Fprintln(os.Stderr, err)
}

// testing flow
func Test(t *engine.Test) {
	fmt.Printf("==> %s\n", t.Name)
}

func Command(_ *engine.Test, cmd engine.Command) {
	fmt.Printf("--> %s\n", cmd)
}

func TestResult(result engine.TestResult, err error) {
	if err != nil {
		fmt.Println("ERR", err)
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
			fmt.Printf("%8s: %s\n", chk.Name, line)
		}
	}
}
