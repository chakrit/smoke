package p

import "github.com/chakrit/smoke/engine"

type Hooks struct{}

var _ engine.RunHooks = Hooks{}

func (Hooks) BeforeTest(t *engine.Test) {
	Test(t)
}

func (Hooks) BeforeCommand(t *engine.Test, cmd engine.Command) {
	Command(t, cmd)
}

func (Hooks) AfterCommand(t *engine.Test, cmd engine.Command, result engine.CommandResult, err error) {
	CommandResult(result, err)
}

func (Hooks) AfterTest(t *engine.Test, result engine.TestResult, err error) {
	TestResult(result, err)
}
