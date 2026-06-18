package testspecs

import (
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
		if parent.Name != "" {
			t.Name = parent.Name + ` \ ` + t.Name
		}
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
			t.Name = t.Filename
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

// Tests flattens the resolved spec tree into engine.Tests. It splits into a
// total parse (build the value-or-error IR, never fails) and a validate fold
// (collect every error across the tree). On any error it returns them all
// aggregated, in depth-first spec order; see test_ir.go.
func (t *TestSpec) Tests() ([]*engine.Test, error) {
	return validate(parse(t))
}
