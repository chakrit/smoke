package testspecs

import (
	"errors"
	"strings"

	"github.com/chakrit/smoke/checks"
	"github.com/chakrit/smoke/engine"
)

// parsed is the value-or-error carrier: a total parse stores either the parsed
// value or the error explaining why the field could not parse. It never panics
// and is never nil-as-signal — validation reads err explicitly.
type parsed[T any] struct {
	val T
	err error
}

func parseOK[T any](v T) parsed[T]      { return parsed[T]{val: v} }
func parseErr[T any](e error) parsed[T] { return parsed[T]{err: e} }

// testIR is one node of the intermediate representation: either a node that
// builds an engine.Test (commands present) or a command-less leaf that only
// carries an error. The tree is flattened in depth-first spec order so the
// validate fold reports problems in the order the author wrote them.
type testIR interface {
	// errs returns every parse error this node carries, in field order.
	errs() []error
	// build materializes the engine.Test, or false when this node contributes
	// none (a command-less leaf). Only called after validation finds no errors.
	build() (*engine.Test, bool)
}

// buildableTest is a node with commands: its checks and run config are parsed
// into carriers so a bad check name or timeout becomes data, not an early abort.
type buildableTest struct {
	name     string
	commands []engine.Command
	checks   []parsed[checks.Interface]
	runCfg   parsed[*engine.Config]
}

func (b buildableTest) errs() []error {
	var out []error
	for _, c := range b.checks {
		if c.err != nil {
			out = append(out, c.err)
		}
	}
	if b.runCfg.err != nil {
		out = append(out, b.runCfg.err)
	}
	return out
}

func (b buildableTest) build() (*engine.Test, bool) {
	allchecks := make([]checks.Interface, len(b.checks))
	for i, c := range b.checks {
		allchecks[i] = c.val
	}
	return &engine.Test{
		Name:      b.name,
		RunConfig: b.runCfg.val,
		Commands:  b.commands,
		Checks:    allchecks,
	}, true
}

// leafError is a node with neither commands nor children — a malformed spec.
type leafError struct {
	err error
}

func (l leafError) errs() []error               { return []error{l.err} }
func (l leafError) build() (*engine.Test, bool) { return nil, false }

// parse is the total pass: it never fails. It walks the resolved TestSpec tree
// in depth-first order and builds the IR, deferring every failable decision
// into a carrier.
func parse(t *TestSpec) []testIR {
	var ir []testIR

	switch {
	case len(t.Commands) > 0:
		ir = append(ir, parseBuildable(t))
	case len(t.Children) == 0:
		ir = append(ir, leafError{
			err: errors.New("test `" + t.Name + "` is a leaf with no commands"),
		})
	}

	for _, child := range t.Children {
		ir = append(ir, parse(child)...)
	}
	return ir
}

func parseBuildable(t *TestSpec) buildableTest {
	commands := make([]engine.Command, len(t.Commands))
	for i, cmdstr := range t.Commands {
		commands[i] = engine.Command(strings.TrimSpace(cmdstr))
	}

	parsedChecks := make([]parsed[checks.Interface], len(t.Checks))
	for i, name := range t.Checks {
		if check := checks.Parse(name); check != nil {
			parsedChecks[i] = parseOK(check)
		} else {
			parsedChecks[i] = parseErr[checks.Interface](errors.New("unknown check `" + name + "`"))
		}
	}

	var runCfg parsed[*engine.Config]
	if cfg, err := t.Config.RunConfig(); err != nil {
		runCfg = parseErr[*engine.Config](err)
	} else {
		runCfg = parseOK(cfg)
	}

	return buildableTest{
		name:     t.Name,
		commands: commands,
		checks:   parsedChecks,
		runCfg:   runCfg,
	}
}

// validate folds the IR collecting errors across the whole tree in spec order.
// All-errors vs first-error is the one-line change at the marked break.
func validate(ir []testIR) ([]*engine.Test, error) {
	var tests []*engine.Test
	var errs []error

	for _, node := range ir {
		if nodeErrs := node.errs(); len(nodeErrs) > 0 {
			errs = append(errs, nodeErrs...)
			continue // first-error mode: replace with `break`.
		}
		if test, ok := node.build(); ok {
			tests = append(tests, test)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return tests, nil
}
