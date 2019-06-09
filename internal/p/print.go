package p

import (
	"fmt"
	"os"
	"strings"

	"github.com/chakrit/smoke/engine"
)

func UsageHint(s string) {
	fmt.Println(s)
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
		fmt.Printf("\t%s:\n", strings.TrimSpace(string(cmd.Command)))
		for _, chk := range cmd.Checks {
			fmt.Printf("\t\t%s:\n\t\t---\n", chk.Name)

			lines := strings.Split(string(chk.Data), "\n")
			for _, line := range lines {
				fmt.Printf("\t\t\t%s\n", line)
			}

			fmt.Printf("\t\t---\n")
		}
	}
}
