package engine

import "slices"

// Pattern is one --include/--exclude token: a substring matched against a
// TestName's flattened path.
type Pattern string

// Filter is the test selection expressed by --include/--exclude. A name is
// selected when it matches some include (or there are none) and matches no
// exclude.
type Filter struct {
	Includes []Pattern
	Excludes []Pattern
}

// NewFilter is the CLI boundary: it mints Patterns from the raw flag strings. The
// one place string becomes Pattern, mirroring how TestName is minted only at the
// flatten gate.
func NewFilter(includes, excludes []string) Filter {
	return Filter{
		Includes: toPatterns(includes),
		Excludes: toPatterns(excludes),
	}
}

func toPatterns(strs []string) []Pattern {
	if len(strs) == 0 {
		return nil
	}
	patterns := make([]Pattern, len(strs))
	for i, s := range strs {
		patterns[i] = Pattern(s)
	}
	return patterns
}

// Active reports whether the filter narrows the run at all — i.e. whether this is
// a partial run rather than the whole suite.
func (f Filter) Active() bool {
	return len(f.Includes) > 0 || len(f.Excludes) > 0
}

func (f Filter) Selects(n TestName) bool {
	if len(f.Includes) > 0 && !n.matchesAny(f.Includes) {
		return false
	}
	return !n.matchesAny(f.Excludes)
}

func (n TestName) matchesAny(patterns []Pattern) bool {
	return slices.ContainsFunc(patterns, n.Matches)
}

// Select keeps the items the filter selects, preserving order. The name extractor
// hands over each item's TestName without a cast.
func Select[T any](f Filter, items []T, name func(T) TestName) []T {
	var selected []T
	for _, item := range items {
		if f.Selects(name(item)) {
			selected = append(selected, item)
		}
	}
	return selected
}
