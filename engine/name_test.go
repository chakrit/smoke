package engine

import "testing"

func TestTestNameChild(t *testing.T) {
	cases := []struct {
		name    string
		parent  TestName
		segment string
		want    TestName
	}{
		{"root seed from empty", "", "spec.yml", "spec.yml"},
		{"single child", "spec.yml", "build", `spec.yml \ build`},
		{"nested child", `spec.yml \ build`, "linux", `spec.yml \ build \ linux`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.parent.Child(c.segment); got != c.want {
				t.Errorf("Child(%q) on %q = %q, want %q", c.segment, c.parent, got, c.want)
			}
		})
	}
}

func TestTestNameMatches(t *testing.T) {
	const name TestName = `spec.yml \ build \ linux`

	cases := []struct {
		pattern Pattern
		want    bool
	}{
		{"build", true},    // interior segment
		{"linux", true},    // leaf segment
		{"spec.yml", true}, // root segment
		{` \ `, true},      // the separator is part of the path
		{"windows", false}, // absent
		{"", true},         // empty pattern matches everything
	}

	for _, c := range cases {
		if got := name.Matches(c.pattern); got != c.want {
			t.Errorf("%q.Matches(%q) = %v, want %v", name, c.pattern, got, c.want)
		}
	}
}
