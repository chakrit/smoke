package main

import "testing"

func TestLockFilename(t *testing.T) {
	cases := map[string]string{
		"foo.yml":   "foo.lock.yml",
		"foo.yaml":  "foo.lock.yaml",
		"foo.cue":   "foo.lock.yml",
		"foo.json":  "foo.lock.yml",
		"foo.jsonl": "foo.lock.yml",
	}

	for in, want := range cases {
		if got := lockFilename(in); got != want {
			t.Errorf("lockFilename(%q) = %q, want %q", in, got, want)
		}
	}
}
