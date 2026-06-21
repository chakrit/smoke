package testspecs

import (
	"fmt"
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
	return t.tests("")
}

func (t *TestSpec) tests(parent engine.TestName) (tests []*engine.Test, err error) {
	name := parent.Child(t.Name)

	if len(t.Children) == 0 && len(t.Commands) == 0 {
		return nil, fmt.Errorf("test `%s` is a leaf with no commands", name)
	}

	if len(t.Commands) > 0 {
		var commands []engine.Command
		for _, cmdstr := range t.Commands {
			cmdstr = strings.TrimSpace(cmdstr)
			commands = append(commands, engine.Command(cmdstr))
		}

		var allchecks []checks.Interface
		for _, checkname := range t.Checks {
			if check := checks.Parse(checkname); check == nil {
				return nil, fmt.Errorf("unknown check `%s`", checkname)
			} else {
				allchecks = append(allchecks, check)
			}
		}

		runcfg, err := t.Config.RunConfig()
		if err != nil {
			return nil, err
		}

		tests = append(tests, &engine.Test{
			Name:      name,
			RunConfig: runcfg,
			Commands:  commands,
			Checks:    allchecks,
		})
	}

	for _, subt := range t.Children {
		if subtests, err := subt.tests(name); err != nil {
			return nil, err
		} else {
			tests = append(tests, subtests...)
		}
	}

	return tests, nil
}
