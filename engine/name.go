package engine

import "strings"

// TestName is the flattened, hierarchical identity of a test — a path of
// segments joined by nameSep. It is minted only by the spec flatten walk
// (testspecs), where the full set is checked for uniqueness; everywhere else it
// is received and used, never re-derived from a raw string.
type TestName string

const nameSep TestName = ` \ `

// Child composes a sub-name under the receiver. An empty receiver seeds the root
// (the first segment stands alone); otherwise the segment is appended after the
// separator.
func (n TestName) Child(segment string) TestName {
	if n == "" {
		return TestName(segment)
	}
	return n + nameSep + TestName(segment)
}

func (n TestName) String() string { return string(n) }

// Matches reports whether the pattern selects this name. Substring containment
// against the flattened path — the same semantics the CLI --include/--exclude
// filter has always used.
func (n TestName) Matches(p Pattern) bool {
	return strings.Contains(string(n), string(p))
}
