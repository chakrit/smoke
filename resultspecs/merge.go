package resultspecs

import "github.com/chakrit/smoke/engine"

// Merge rebuilds a lock for a partial commit. It walks the spec order: each name
// takes its fresh result if it was run this pass, otherwise carries forward the
// existing lock entry. Names in neither (never committed) are dropped, as are
// names absent from the spec order (deleted tests) — matching a full commit.
//
// Order follows the spec, not the lock, because compare is order-sensitive: a
// newly committed test must land in its spec position, not be appended.
func Merge(order []engine.TestName, fresh, existing []TestResultSpec) []TestResultSpec {
	freshByName := indexByName(fresh)
	carriedByName := indexByName(existing)

	var merged []TestResultSpec
	for _, name := range order {
		if spec, ok := freshByName[name]; ok {
			merged = append(merged, spec)
		} else if spec, ok := carriedByName[name]; ok {
			merged = append(merged, spec)
		}
	}
	return merged
}

func indexByName(specs []TestResultSpec) map[engine.TestName]TestResultSpec {
	byName := make(map[engine.TestName]TestResultSpec, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
	}
	return byName
}
