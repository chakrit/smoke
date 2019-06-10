package engine

import (
	"io"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"

	"github.com/chakrit/smoke/checks"
)

type (
	CommandResult struct {
		Command Command `yaml:"command"`
		Err     error
		Checks  []checks.Output
	}

	TestResult struct {
		Test     *Test `yaml:"test"`
		Err      error
		Commands []CommandResult
	}
)

func (r CommandResult) Save(w io.Writer) error {
	return errors.Wrap(yaml.NewEncoder(w).Encode(r), "save")
}

func (r TestResult) Save(w io.Writer) error {
	return errors.Wrap(yaml.NewEncoder(w).Encode(r), "save")
}

func SaveAll(w io.Writer, results []TestResult) error {
	return errors.Wrap(yaml.NewEncoder(w).Encode(results), "save")
}
