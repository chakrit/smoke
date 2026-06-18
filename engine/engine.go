package engine

import (
	"github.com/chakrit/smoke/checks"
)

type (
	Command string

	// TestID is the identity of a test across a run and its lock: the value a
	// lock merge keys on to decide which entry a result replaces. Derived from
	// the flattened name today; centralized here so the derivation can grow
	// richer without touching call sites.
	TestID string

	Test struct {
		Name      string
		RunConfig *Config

		Commands []Command
		Checks   []checks.Interface
	}

	CommandResult struct {
		Command Command `yaml:"command"`
		Checks  []checks.Result
	}

	TestResult struct {
		Test     *Test `yaml:"test"`
		Commands []CommandResult
	}
)
