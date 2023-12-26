package testspecs

import (
	"time"

	"github.com/chakrit/smoke/engine"
)

type ConfigSpec struct {
	WorkDir     string         `yaml:"workdir"`
	Env         []string       `yaml:"env"`
	Interpreter string         `yaml:"interpreter"`
	Timeout     *time.Duration `yaml:"timeout"`
}

// Resolve() applies parent-child value overriding and extension logic.
func (c *ConfigSpec) Resolve(parent *ConfigSpec) {
	def := engine.DefaultConfig

	if parent != nil {
		c.WorkDir = resolvePaths(parent.WorkDir, c.WorkDir)
		c.Env = append(parent.Env, c.Env...)
		c.Interpreter = resolveStrings(c.Interpreter, parent.Interpreter)
		c.Timeout = resolveDurations(c.Timeout, parent.Timeout)
	} else {
		c.WorkDir = resolvePaths(def.WorkDir, c.WorkDir)
		c.Env = append(def.Env, c.Env...)
		c.Interpreter = resolveStrings(c.Interpreter, def.Interpreter)
		c.Timeout = resolveDurations(c.Timeout, &def.Timeout)
	}
}

func (c *ConfigSpec) RunConfig() (*engine.Config, error) {
	if c == nil {
		return nil, nil
	}

	runcfg := &engine.Config{
		WorkDir:     c.WorkDir,
		Env:         c.Env,
		Interpreter: c.Interpreter,
	}
	if c.Timeout == nil {
		runcfg.Timeout = engine.DefaultConfig.Timeout
	} else {
		runcfg.Timeout = *c.Timeout
	}

	return runcfg, nil
}
