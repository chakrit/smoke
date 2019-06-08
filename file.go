package main

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type File struct {
	Filename string
	Tests    []*Test `yaml:"tests"`
}

func Load(filename string) (*File, error) {
	infile, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "i/o error")
	}
	defer infile.Close()

	outfile := &File{}
	if err := yaml.NewDecoder(infile).Decode(outfile); err != nil {
		return nil, err
	}

	outfile.Filename = filename
	return outfile, nil
}

func (f *File) Run() (*Result, error) {
	for _, test := range f.Tests {
		// TODO: Multi-error
		if err := test.Run(); err != nil {
			return err
		}
	}
}
