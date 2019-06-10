package specs

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func Load(filename string) (*Test, error) {
	infile, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "i/o error")
	}
	defer infile.Close()

	root := &Test{}
	if err := yaml.NewDecoder(infile).Decode(root); err != nil {
		return nil, err
	}

	root.Filename = filename
	root.resolve(nil)
	return root, nil
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
