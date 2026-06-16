package testspecs

import (
	"io"
	"path/filepath"

	"github.com/chakrit/smoke/engine"
)

func Load(reader io.Reader, filename string) ([]*engine.Test, error) {
	loader, err := loaderFor(filename)
	if err != nil {
		return nil, err
	}

	root, err := loader.Load(reader)
	if err != nil {
		return nil, err
	}

	root.Filename = filename
	root.Resolve(nil)
	return root.Tests()
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
