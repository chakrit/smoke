package testspecs

// import (
// 	"fmt"
// 	"io"
// 	"time"

// 	"cuelang.org/go/cue"
// 	"cuelang.org/go/cue/cuecontext"
// )

//go:embed schema.cue
var schemaCueContent string

// type TestSpec struct {
// 	Name     string      `json:"name"`
// 	Filename string      `json:"-"`
// 	Children []*TestSpec `json:"tests"`

// 	Image       string            `json:"image"`
// 	Interpreter string            `json:"interpreter"`
// 	WorkDir     string            `json:"workdir"`
// 	Timeout     time.Duration     `json:"timeout"`
// 	Env         map[string]string `json:"env"`

// 	BeforeCommands []string `json:"before"`
// 	AfterCommands  []string `json:"after"`
// 	Commands       []string `json:"commands"`
// 	Checks         []string `json:"checks"`
// }

// func Load(reader io.Reader, filename string) (*TestSpec, error) {
// 	cuectx := cuecontext.New()
// 	value := cuectx.CompileString(schemaCueContent, cue.Filename("smoke-builtins/schema.cue"))
// 	if value.Err() != nil {
// 		return nil, wrapErr(value.Err())
// 	}

// 	content, err := io.ReadAll(reader)
// 	if err != nil {
// 		return nil, err
// 	}

// 	value = cuectx.CompileBytes(content, cue.Filename(name), cue.Scope(value))
// 	if value.Err() != nil {
// 		return nil, wrapErr(value.Err())
// 	}

// 	root := &TestSpec{}
// 	if err := value.Decode(root); err != nil {
// 		return nil, wrapErr(err)
// 	}

// 	root.Filename = filename
// 	root.wakeUp(nil)
// 	return root, nil
// }

// func (s *TestSpec) wakeUp(parent *TestSpec) {
// 	if s.Name == "" {
// 		s.Name = "(unnamed)"
// 	}

// 	if parent != nil {
// 		s.Name = parent.Name + " / " + s.Name
// 		s.Filename = parent.Filename

// 		s.BeforeCommands = append(parent.BeforeCommands, s.BeforeCommands...)
// 		s.AfterCommands = append(parent.AfterCommands, s.AfterCommands...)
// 		s.Checks = append(parent.Checks, s.Checks...)
// 	}

// 	for _, child := range s.Children {
// 		child.wakeUp(s)
// 	}
// }

// func wrapErr(err error) error {
// 	return fmt.Errorf("testspecs: %w", err)
// }
