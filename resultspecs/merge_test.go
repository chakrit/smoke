package resultspecs

import (
	"testing"

	"github.com/chakrit/smoke/engine"
)

// spec tags a result with a marker command so a fresh entry is distinguishable
// from the existing one it replaces.
func spec(name engine.TestName, marker string) TestResultSpec {
	return TestResultSpec{
		Name:     name,
		Commands: []CommandResultSpec{{Command: marker}},
	}
}

func markers(specs []TestResultSpec) map[engine.TestName]string {
	out := make(map[engine.TestName]string, len(specs))
	for _, s := range specs {
		out[s.Name] = s.Commands[0].Command
	}
	return out
}

func TestMerge(t *testing.T) {
	const (
		a engine.TestName = "spec \\ a"
		b engine.TestName = "spec \\ b"
		c engine.TestName = "spec \\ c"
	)

	t.Run("filtered run carries forward unselected entries in spec order", func(t *testing.T) {
		order := []engine.TestName{a, b, c}
		existing := []TestResultSpec{spec(a, "old-a"), spec(b, "old-b"), spec(c, "old-c")}
		fresh := []TestResultSpec{spec(b, "new-b")}

		got := Merge(order, fresh, existing)

		if want := []engine.TestName{a, b, c}; !sameOrder(got, want) {
			t.Fatalf("order = %v, want %v", names(got), want)
		}
		m := markers(got)
		if m[a] != "old-a" || m[c] != "old-c" {
			t.Errorf("unselected entries not carried forward: %v", m)
		}
		if m[b] != "new-b" {
			t.Errorf("selected entry not refreshed: b = %q, want new-b", m[b])
		}
	})

	t.Run("empty lock writes only the fresh entries", func(t *testing.T) {
		order := []engine.TestName{a, b, c}
		fresh := []TestResultSpec{spec(b, "new-b")}

		got := Merge(order, fresh, nil)

		if want := []engine.TestName{b}; !sameOrder(got, want) {
			t.Fatalf("order = %v, want %v (never-committed tests stay absent)", names(got), want)
		}
	})

	t.Run("entry gone from the spec order is dropped", func(t *testing.T) {
		order := []engine.TestName{b, c} // a removed from the spec
		existing := []TestResultSpec{spec(a, "old-a"), spec(b, "old-b"), spec(c, "old-c")}
		fresh := []TestResultSpec{spec(b, "new-b")}

		got := Merge(order, fresh, existing)

		if want := []engine.TestName{b, c}; !sameOrder(got, want) {
			t.Fatalf("order = %v, want %v (deleted test dropped)", names(got), want)
		}
	})
}

func names(specs []TestResultSpec) []engine.TestName {
	out := make([]engine.TestName, len(specs))
	for i, s := range specs {
		out[i] = s.Name
	}
	return out
}

func sameOrder(got []TestResultSpec, want []engine.TestName) bool {
	gotNames := names(got)
	if len(gotNames) != len(want) {
		return false
	}
	for i := range want {
		if gotNames[i] != want[i] {
			return false
		}
	}
	return true
}
