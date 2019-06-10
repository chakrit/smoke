package testspecs

import (
	"time"

	lib "github.com/chakrit/smoke/engine"
)

type ConfigSpec struct {
	WorkDir     string         `yaml:"workdir"`
	Env         []string       `yaml:"env"`
	Interpreter string         `yaml:"interpreter"`
	Timeout     *time.Duration `yaml:"timeout"`
}

// Resolve() applies parent-child value overriding and extension logic.
func (c *ConfigSpec) Resolve(parent *ConfigSpec) {
	def := lib.DefaultConfig

	if parent != nil {
		// TODO: This should instead be relative, not just a first non-empty.
		//   i.e. when parent has workdir = "somefolder" and child has
		//   workdir = "..", this should result in the current folder being
		//   workdir
		c.WorkDir = resolveStrings(c.WorkDir, parent.WorkDir)
		c.Env = append(parent.Env, c.Env...)
		c.Interpreter = resolveStrings(c.Interpreter, parent.Interpreter)
		c.Timeout = resolveDurations(c.Timeout, parent.Timeout)
	} else {
		c.WorkDir = resolveStrings(c.WorkDir, def.WorkDir)
		c.Env = append(def.Env, c.Env...)
		c.Interpreter = resolveStrings(c.Interpreter, def.Interpreter)
		c.Timeout = resolveDurations(c.Timeout, &def.Timeout)
	}
}

func (c *ConfigSpec) RunConfig() (*lib.Config, error) {
	if c == nil {
		return nil, nil
	}

	runcfg := &lib.Config{
		WorkDir:     c.WorkDir,
		Env:         c.Env,
		Interpreter: c.Interpreter,
	}
	if c.Timeout == nil {
		runcfg.Timeout = lib.DefaultConfig.Timeout
	} else {
		runcfg.Timeout = *c.Timeout
	}

	return runcfg, nil
}
