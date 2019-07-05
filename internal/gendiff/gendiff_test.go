package gendiff

import (
	"fmt"
	"testing"

	r "github.com/stretchr/testify/require"
)

type testcase struct {
	left  string
	right string
	diff  []Diff
}

var _ Interface = testcase{}

func (c testcase) LeftLen() int        { return len(c.left) }
func (c testcase) RightLen() int       { return len(c.right) }
func (c testcase) Equal(l, r int) bool { return c.left[l] == c.right[r] }

func (c testcase) Name() string {
	return fmt.Sprintf("Diff %#v against %#v result in %d edits",
		c.left, c.right, len(c.diff))
}

var cases = []testcase{
	{"a", "", []Diff{
		{Delete, 0, 1, 0, 0},
	}},
	{"", "a", []Diff{
		{Insert, 0, 0, 0, 1},
	}},
	{"aaa", "bbb", []Diff{
		{Delete, 0, 3, 0, 0},
		{Insert, 3, 3, 0, 3},
	}},
	{"abce", "acde", []Diff{
		{Match, 0, 1, 0, 1},
		{Delete, 1, 2, 1, 1},
		{Match, 2, 3, 1, 2},
		{Insert, 3, 3, 2, 3},
		{Match, 3, 4, 3, 4},
	}},
	{"aaabbbccceee", "aaacccdddeee", []Diff{
		{Match, 0, 3, 0, 3},
		{Delete, 3, 6, 3, 3},
		{Match, 6, 9, 3, 6},
		{Insert, 9, 9, 6, 9},
		{Match, 9, 12, 9, 12},
	}},
	{"bbbcccddd", "ccceee", []Diff{
		{Delete, 0, 3, 0, 0},
		{Match, 3, 6, 0, 3},
		{Delete, 6, 9, 3, 3},
		{Insert, 9, 9, 3, 6},
	}},
}

func TestMake(t *testing.T) {
	for _, test := range cases {
		t.Run(test.Name(), func(tt *testing.T) {
			r.Equal(tt, test.diff, Make(test))
		})
	}
}
