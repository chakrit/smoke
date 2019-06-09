package specs

import (
	"time"

	lib "github.com/chakrit/smoke/engine"
)

type Config struct {
	WorkDir     string         `yaml:"workdir"`
	Env         []string       `yaml:"env"`
	Interpreter string         `yaml:"interpreter"`
	Timeout     *time.Duration `yaml:"timeout"`
}

func (c *Config) resolve(parent *Config) {
	def := lib.DefaultConfig

	if parent != nil {
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

func (n *Config) RunConfig() (*lib.Config, error) {
	if n == nil {
		return nil, nil
	}

	runcfg := &lib.Config{
		WorkDir:     n.WorkDir,
		Env:         n.Env,
		Interpreter: n.Interpreter,
	}
	if n.Timeout == nil {
		runcfg.Timeout = lib.DefaultConfig.Timeout
	} else {
		runcfg.Timeout = *n.Timeout
	}

	return runcfg, nil
}
