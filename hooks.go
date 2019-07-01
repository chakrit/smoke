package main

import (
	"github.com/chakrit/smoke/engine"
	"github.com/chakrit/smoke/internal/p"
)

type Hooks struct {
	WrapErr func(t *engine.Test, err error) error
}

var _ engine.RunHooks = Hooks{}

func (h Hooks) BeforeTest(t *engine.Test) {
	p.Test(t)
}

func (h Hooks) BeforeCommand(t *engine.Test, cmd engine.Command) {
	p.Command(t, cmd)
}

func (h Hooks) AfterCommand(t *engine.Test, cmd engine.Command, result engine.CommandResult, err error) {
	p.CommandResult(result, h.wrapErr(t, err))
}

func (h Hooks) AfterTest(t *engine.Test, result engine.TestResult, err error) {
	p.TestResult(result, h.wrapErr(t, err))
}

func (h Hooks) wrapErr(t *engine.Test, err error) error {
	if err != nil && h.WrapErr != nil {
		return h.WrapErr(t, err)
	} else {
		return err
	}
}
