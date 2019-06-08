package main

import "github.com/pkg/errors"

type Runnable interface {
	Run(*Result, error)
}

type Result struct {
	Name       string
	Exit       int
	Stdout     []byte
	Stderr     []byte
	Subresults []*Result
}

func RunAll(name string, tests []*Test) (*Result, error) {
	return nil, errors.New("not implemented")

}
