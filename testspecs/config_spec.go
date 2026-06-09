package testspecs

import (
	"fmt"
	"time"

	"github.com/chakrit/smoke/engine"
)

type ConfigSpec struct {
	WorkDir     string   `yaml:"workdir" json:"workdir"`
	Env         []string `yaml:"env" json:"env"`
	Interpreter string   `yaml:"interpreter" json:"interpreter"`
	Timeout     string   `yaml:"timeout" json:"timeout"`
}

// Resolve() applies parent-child value overriding and extension logic.
func (c *ConfigSpec) Resolve(parent *ConfigSpec) {
	def := engine.DefaultConfig

	if parent != nil {
		c.WorkDir = resolvePaths(parent.WorkDir, c.WorkDir)
		c.Env = append(parent.Env, c.Env...)
		c.Interpreter = resolveStrings(c.Interpreter, parent.Interpreter)
		c.Timeout = resolveStrings(c.Timeout, parent.Timeout)
	} else {
		c.WorkDir = resolvePaths(def.WorkDir, c.WorkDir)
		c.Env = append(def.Env, c.Env...)
		c.Interpreter = resolveStrings(c.Interpreter, def.Interpreter)
		c.Timeout = resolveStrings(c.Timeout, def.Timeout.String())
	}
}

func (c *ConfigSpec) RunConfig() (*engine.Config, error) {
	if c == nil {
		return nil, nil
	}

	timeout := engine.DefaultConfig.Timeout
	if c.Timeout != "" {
		parsed, err := time.ParseDuration(c.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout %q: %w", c.Timeout, err)
		}
		timeout = parsed
	}

	return &engine.Config{
		WorkDir:     c.WorkDir,
		Env:         c.Env,
		Interpreter: c.Interpreter,
		Timeout:     timeout,
	}, nil
}
