package engine

import "time"

type Config struct {
	WorkDir     string
	Env         []string
	Interpreter string
	Timeout     time.Duration
}

var DefaultConfig = &Config{
	Env:         nil,
	Interpreter: "/bin/bash",
	Timeout:     3 * time.Second,
}

func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}
