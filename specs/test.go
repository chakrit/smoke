package specs

import (
	lib "smoke/smokelib"

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
	var commands []lib.Command
	for _, cmdstr := range t.Commands {
		commands = append(commands, lib.Command(cmdstr))
	}

	var checks []lib.Check
	for _, checkstr := range t.Checks {
		switch checkstr {
		case "stdout":
			checks = append(checks, lib.StdoutCheck)
		case "stderr":
			checks = append(checks, lib.StderrCheck)
		case "exitcode":
			checks = append(checks, lib.ExitCodeCheck)
		default:
			return nil, errors.WithMessage(lib.ErrCheckSpec,
				"`"+checkstr+"`")
		}
	}

	runcfg, err := t.Config.RunConfig()
	if err != nil {
		return nil, err
	}

	if len(t.Commands) > 0 {
		tests = append(tests, &lib.Test{
			Name:      t.Name,
			RunConfig: runcfg,
			Commands:  commands,
			Checks:    checks,
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
