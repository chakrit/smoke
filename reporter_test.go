package main

import "testing"

// status.Merge folds per-spec verdicts into the aggregate exit. The load-bearing
// property: UNCHANGED is the identity, so a later clean spec can never clear an
// earlier drift (the masking bug). Among non-clean specs, the later one wins.
func TestStatusMerge(t *testing.T) {
	cases := []struct {
		name      string
		acc, next status
		want      status
	}{
		{"clean stays clean", statusUnchanged, statusUnchanged, statusUnchanged},
		{"drift over clean accumulator", statusUnchanged, statusChanged, statusChanged},
		{"new over clean accumulator", statusUnchanged, statusNew, statusNew},
		{"clean never clears changed", statusChanged, statusUnchanged, statusChanged},
		{"clean never clears new", statusNew, statusUnchanged, statusNew},
		{"last non-clean wins: new after changed", statusChanged, statusNew, statusNew},
		{"last non-clean wins: changed after new", statusNew, statusChanged, statusChanged},
	}

	for _, c := range cases {
		if got := c.acc.Merge(c.next); got != c.want {
			t.Errorf("%s: %v.Merge(%v) = %v, want %v", c.name, c.acc, c.next, got, c.want)
		}
	}
}
