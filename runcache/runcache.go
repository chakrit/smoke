package runcache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/chakrit/smoke/resultspecs"
	"gopkg.in/yaml.v3"
)

// Snapshot is a persisted run: the observed results plus the provenance needed
// to commit them later without re-running. SpecHash binds the snapshot to the
// exact spec bytes it observed, so a later commit can refuse a spec that has
// since changed. Partial records whether the run was filtered, so that commit
// merges a subset onto the lock rather than overwriting it.
type Snapshot struct {
	SpecHash string                       `yaml:"spec_hash"`
	Partial  bool                         `yaml:"partial"`
	Results  []resultspecs.TestResultSpec `yaml:"results"`
}

// HashSpec is the provenance fingerprint of a spec's bytes.
func HashSpec(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// Save writes the snapshot to the per-spec cache slot. The cache is a
// convenience, never the source of truth — callers treat a failure as "no
// cache", not as a run failure.
func Save(specPath string, snap Snapshot) error {
	path, err := cachePath(specPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return yaml.NewEncoder(file).Encode(snap)
}

// Load reads the snapshot for a spec, returning ok=false when none exists yet.
func Load(specPath string) (snap Snapshot, ok bool, err error) {
	path, err := cachePath(specPath)
	if err != nil {
		return Snapshot{}, false, err
	}

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return Snapshot{}, false, nil
	}
	if err != nil {
		return Snapshot{}, false, err
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&snap); err != nil {
		return Snapshot{}, false, err
	}
	return snap, true, nil
}

// cachePath maps a spec to a stable slot under the user cache dir, keyed by the
// absolute spec path so distinct specs never collide.
func cachePath(specPath string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(specPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "smoke", HashSpec([]byte(abs))+".yml"), nil
}
