package specs

import (
	"os"

	lib "github.com/chakrit/smoke/smokelib"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type File struct {
	Filename string
	RootNode *Test
}

func Load(filename string) (*File, error) {
	infile, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "i/o error")
	}
	defer infile.Close()

	root := &Test{}
	if err := yaml.NewDecoder(infile).Decode(root); err != nil {
		return nil, err
	}

	return &File{
		Filename: filename,
		RootNode: root,
	}, nil
}

func (f *File) Tests() ([]*lib.Test, error) {
	f.RootNode.resolve(nil)
	return f.RootNode.Tests()
}
