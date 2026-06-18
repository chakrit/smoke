package runcache

import (
	"path/filepath"
	"testing"

	"github.com/chakrit/smoke/resultspecs"
)

// HashSpec is the provenance key: identical bytes must fingerprint identically
// and any change must diverge, or the staleness guard can't tell a stale run
// from a fresh one.
func TestHashSpec(t *testing.T) {
	a := HashSpec([]byte("commands: [echo hi]\n"))
	again := HashSpec([]byte("commands: [echo hi]\n"))
	changed := HashSpec([]byte("commands: [echo bye]\n"))

	if a != again {
		t.Errorf("same bytes hashed differently: %q vs %q", a, again)
	}
	if a == changed {
		t.Error("changed bytes hashed identically — guard would miss spec edits")
	}
}

// A round-trip must preserve every provenance field; commit-last reads them back
// to decide refuse-vs-commit and merge-vs-overwrite.
func TestSaveLoadRoundTrip(t *testing.T) {
	redirectCache(t)
	spec := filepath.Join(t.TempDir(), "spec.yml")

	want := Snapshot{
		SpecHash: "deadbeef",
		Partial:  true,
		Results:  []resultspecs.TestResultSpec{{Name: "A"}, {Name: "B"}},
	}
	if err := Save(spec, want); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, ok, err := Load(spec)
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if got.SpecHash != want.SpecHash || got.Partial != want.Partial {
		t.Errorf("provenance lost: got %+v", got)
	}
	if len(got.Results) != 2 || got.Results[1].Name != "B" {
		t.Errorf("results lost: %+v", got.Results)
	}
}

func TestLoadMissingIsNotAnError(t *testing.T) {
	redirectCache(t)

	_, ok, err := Load(filepath.Join(t.TempDir(), "never-run.yml"))
	if err != nil {
		t.Fatalf("missing cache should not error: %v", err)
	}
	if ok {
		t.Error("missing cache reported ok=true")
	}
}

// redirectCache points os.UserCacheDir at a temp dir for the test (HOME on
// macOS, XDG_CACHE_HOME on Linux), so tests never touch the real cache.
func redirectCache(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
}
