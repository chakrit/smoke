package testspecs

import (
	"io"
	"path/filepath"
	"time"

	"github.com/chakrit/smoke/engine"
	"gopkg.in/yaml.v3"
)

func Load(reader io.Reader, filename string) (engine.Collection, error) {
	root := &TestSpec{}
	if err := yaml.NewDecoder(reader).Decode(root); err != nil {
		return nil, err
	}

	root.Filename = filename
	root.Resolve(nil)
	if tests, err := root.Tests(); err != nil {
		return nil, err
	} else {
		return tests, nil
	}
}

func resolvePaths(strs ...string) string {
	return filepath.Join(strs...)
}

func resolveStrings(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}

func resolveDurations(durations ...*time.Duration) *time.Duration {
	for _, dur := range durations {
		if dur != nil {
			return dur
		}
	}
	return nil
}
