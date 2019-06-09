package print

import (
	"fmt"
	"os"

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
