package specs

import (
	"github.com/chakrit/smoke/checks"
	lib "github.com/chakrit/smoke/engine"
	"github.com/pkg/errors"
)

type Test struct {
	Name     string   `yaml:"name"`
	Config   *Config  `yaml:"config"`
	Commands []string `yaml:"commands"`
	Checks   []string `yaml:"checks"`
	Children []*Test  `yaml:"tests"`
}

// resolve() applies parent-child value overriding and extension logic.
func (t *Test) resolve(parent *Test) {
	if parent != nil {
		if parent.Name != "" {
			t.Name = parent.Name + ` \ ` + t.Name
		}
		if t.Config == nil {
			t.Config = parent.Config
		} else {
			t.Config.resolve(parent.Config)
		}
		t.Commands = append(parent.Commands, t.Commands...)
		t.Checks = append(parent.Checks, t.Checks...)

	} else {
		if t.Config == nil {
			t.Config = &Config{}
			t.Config.resolve(nil)
		}
	}

	for _, child := range t.Children {
		child.resolve(t)
	}
}

func (t *Test) Tests() (tests []*lib.Test, err error) {
	if len(t.Commands) > 0 {
		var commands []lib.Command
		for _, cmdstr := range t.Commands {
			commands = append(commands, lib.Command(cmdstr))
		}

		var allchecks []checks.Interface
		for _, name := range t.Checks {
			if check := checks.Parse(name); check == nil {
				return nil, errors.WithMessage(lib.ErrSpec,
					"unknown check `"+name+"`")
			} else {
				allchecks = append(allchecks, check)
			}
		}

		runcfg, err := t.Config.RunConfig()
		if err != nil {
			return nil, err
		}

		tests = append(tests, &lib.Test{
			Name:      t.Name,
			RunConfig: runcfg,
			Commands:  commands,
			Checks:    allchecks,
		})
	}

	for _, subt := range t.Children {
		if subtests, err := subt.Tests(); err != nil {
			return nil, err
		} else {
			tests = append(tests, subtests...)
		}
	}

	return tests, nil
}
