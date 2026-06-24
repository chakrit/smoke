package testspecs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chakrit/smoke/checks"
	"github.com/chakrit/smoke/engine"
)

type TestSpec struct {
	Name     string `yaml:"name" json:"name"`
	Filename string `json:"-"`

	Config   *ConfigSpec `yaml:"config" json:"config"`
	Commands []string    `yaml:"commands" json:"commands"`
	Checks   []string    `yaml:"checks" json:"checks"`

	// Include splices another spec file in as a child of this node, resolved at
	// load time before Resolve and cleared once spliced. Mutually exclusive with
	// Children. See docs/decisions/2026-06-23-include-import-design.md.
	Include  string      `yaml:"include" json:"include"`
	Children []*TestSpec `yaml:"tests" json:"tests"`
}

// Resolve() applies parent-child value overriding and extension logic.
func (t *TestSpec) Resolve(parent *TestSpec) {
	if parent != nil {
		t.Filename = parent.Filename
		if t.Config == nil {
			t.Config = parent.Config
		} else {
			t.Config.Resolve(parent.Config)
		}
		t.Commands = append(parent.Commands, t.Commands...)
		t.Checks = append(parent.Checks, t.Checks...)

	} else {
		if t.Name == "" {
			// Identity is the spec's basename, not the path as typed, so the same
			// spec keys the lock identically across cwd / `./` / abs-vs-rel forms.
			t.Name = filepath.Base(t.Filename)
		}
		if t.Config == nil {
			t.Config = &ConfigSpec{}
			t.Config.Resolve(nil)
		}
	}

	for _, child := range t.Children {
		child.Resolve(t)
	}
}

// Tests flattens the spec tree into the runnable test list. Identity is composed
// here — each node's TestName is its parent's name extended by its own segment —
// so name composition lives at the flatten gate, not in Resolve.
func (t *TestSpec) Tests() ([]*engine.Test, error) {
	tests, errs := t.tests("")
	return tests, errors.Join(errs...)
}

// expandName interpolates $VAR/${VAR} in the node's name segment against its
// resolved env (KEY=value, last wins). Undefined vars expand to empty — the
// os.Expand/envsubst default — and the source is the spec's declared env only,
// never os.Environ. This is what powers parameterized includes: a shared file's
// names resolve against whatever env each importing node passed down.
func (t *TestSpec) expandName() string {
	if t.Config == nil {
		return t.Name
	}
	env := make(map[string]string, len(t.Config.Env))
	for _, kv := range t.Config.Env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			env[k] = v
		}
	}
	return os.Expand(t.Name, func(key string) string { return env[key] })
}

// tests walks the subtree rooted at t, accumulating every error rather than
// stopping at the first, so one Load surfaces all of a spec's tree-walk faults.
// A node's test is appended only if that node built cleanly; children always
// recurse and merge their errors up. Caller (Tests) joins the slice.
func (t *TestSpec) tests(parent engine.TestName) ([]*engine.Test, []error) {
	name := parent.Child(t.expandName())

	var tests []*engine.Test
	var errs []error

	if len(t.Children) == 0 && len(t.Commands) == 0 {
		errs = append(errs, fmt.Errorf("test `%s` is a leaf with no commands", name))
	}

	if len(t.Commands) > 0 {
		var commands []engine.Command
		for _, cmdstr := range t.Commands {
			cmdstr = strings.TrimSpace(cmdstr)
			commands = append(commands, engine.Command(cmdstr))
		}

		var allchecks []checks.Interface
		var checkErrs []error
		for _, checkname := range t.Checks {
			check := checks.Parse(checkname)
			if check == nil {
				checkErrs = append(checkErrs, fmt.Errorf("unknown check `%s`", checkname))
				continue
			}
			allchecks = append(allchecks, check)
		}
		errs = append(errs, checkErrs...)

		runcfg, cfgErr := t.Config.RunConfig()
		if cfgErr != nil {
			errs = append(errs, cfgErr)
		}

		if len(checkErrs) == 0 && cfgErr == nil {
			tests = append(tests, &engine.Test{
				Name:      name,
				RunConfig: runcfg,
				Commands:  commands,
				Checks:    allchecks,
			})
		}
	}

	for _, subt := range t.Children {
		subtests, suberrs := subt.tests(name)
		tests = append(tests, subtests...)
		errs = append(errs, suberrs...)
	}

	return tests, errs
}
