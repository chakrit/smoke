package engine

import (
	"slices"
	"testing"
)

func TestFilterSelects(t *testing.T) {
	const (
		build TestName = `spec.yml \ build`
		test  TestName = `spec.yml \ test`
		lint  TestName = `spec.yml \ lint`
	)

	cases := []struct {
		name     string
		filter   Filter
		selected []TestName
		rejected []TestName
	}{
		{
			name:     "no filter selects all",
			filter:   Filter{},
			selected: []TestName{build, test, lint},
		},
		{
			name:     "include narrows to matches",
			filter:   Filter{Includes: []Pattern{"build", "lint"}},
			selected: []TestName{build, lint},
			rejected: []TestName{test},
		},
		{
			name:     "exclude removes matches",
			filter:   Filter{Excludes: []Pattern{"test"}},
			selected: []TestName{build, lint},
			rejected: []TestName{test},
		},
		{
			name:     "exclude wins over include",
			filter:   Filter{Includes: []Pattern{"spec.yml"}, Excludes: []Pattern{"lint"}},
			selected: []TestName{build, test},
			rejected: []TestName{lint},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, n := range c.selected {
				if !c.filter.Selects(n) {
					t.Errorf("Selects(%q) = false, want true", n)
				}
			}
			for _, n := range c.rejected {
				if c.filter.Selects(n) {
					t.Errorf("Selects(%q) = true, want false", n)
				}
			}
		})
	}
}

func TestFilterActive(t *testing.T) {
	cases := []struct {
		name   string
		filter Filter
		want   bool
	}{
		{"empty is inactive", Filter{}, false},
		{"include is active", Filter{Includes: []Pattern{"x"}}, true},
		{"exclude is active", Filter{Excludes: []Pattern{"x"}}, true},
	}

	for _, c := range cases {
		if got := c.filter.Active(); got != c.want {
			t.Errorf("%s: Active() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestNewFilter(t *testing.T) {
	f := NewFilter([]string{"a", "b"}, []string{"c"})

	if want := []Pattern{"a", "b"}; !slices.Equal(f.Includes, want) {
		t.Errorf("Includes = %v, want %v", f.Includes, want)
	}
	if want := []Pattern{"c"}; !slices.Equal(f.Excludes, want) {
		t.Errorf("Excludes = %v, want %v", f.Excludes, want)
	}
}

func TestSelect(t *testing.T) {
	type item struct {
		id   int
		name TestName
	}
	items := []item{
		{1, `spec.yml \ build`},
		{2, `spec.yml \ test`},
		{3, `spec.yml \ lint`},
	}
	nameOf := func(it item) TestName { return it.name }

	got := Select(Filter{Excludes: []Pattern{"test"}}, items, nameOf)

	want := []int{1, 3}
	var gotIDs []int
	for _, it := range got {
		gotIDs = append(gotIDs, it.id)
	}
	if !slices.Equal(gotIDs, want) {
		t.Errorf("selected ids = %v, want %v", gotIDs, want)
	}
}
