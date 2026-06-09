package checks

import (
	"errors"
	"os/exec"
	"strings"
)

// Timeout is a synthetic check the runner emits when a command exceeds its
// deadline. A timeout is the command misbehaving observably, so it is recorded
// as a check result and compares as drift — see docs/spec/exit-codes.md. It is
// never authored in a test spec, so Prepare/Collect are unreachable.
var Timeout Interface = timeout{}

type timeout struct{}

func (timeout) Spec() string { return "timeout" }

func (timeout) Prepare(*exec.Cmd) error           { return nil }
func (timeout) Collect(*exec.Cmd) ([]byte, error) { return nil, errors.New("timeout: not collectable") }

func (timeout) Format(buf []byte) ([]string, error) {
	if len(buf) == 0 {
		return []string{}, nil
	}
	return strings.Split(string(buf), "\n"), nil
}
