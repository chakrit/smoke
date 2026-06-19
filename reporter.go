package main

import (
	"github.com/chakrit/smoke/internal/p"
	"github.com/chakrit/smoke/resultspecs"
)

// status is the compare outcome — one of three drift states, each owning its
// frozen exit code (docs/spec/exit-codes.md). The state→code mapping lives here
// so no caller hand-pairs a label with a literal.
type status int

const (
	statusUnchanged status = iota
	statusChanged
	statusNew
)

func (s status) String() string {
	switch s {
	case statusChanged:
		return "changed"
	case statusNew:
		return "new"
	default:
		return "unchanged"
	}
}

func (s status) ExitCode() int {
	switch s {
	case statusChanged:
		return p.ExitChanged
	case statusNew:
		return p.ExitNew
	default:
		return p.ExitUnchanged
	}
}

// Merge folds the next spec's verdict into the running aggregate for a multi-spec
// run. UNCHANGED is the identity: a clean spec never clears an earlier drift —
// that would re-mask it, the exact bug aggregation exists to prevent. Among
// non-clean specs the later verdict wins (last-write-win), so the run exits
// non-zero whenever any spec drifted. See docs/spec/exit-codes.md.
func (s status) Merge(next status) status {
	if next == statusUnchanged {
		return s
	}
	return next
}

// A reporter renders one compare outcome. compareResults selects an
// implementation by output format, then exits with status.ExitCode().
type reporter interface {
	Report(lock string, status status, edits []resultspecs.TestEdit) error
}

// consoleReporter renders the human drift report via the p package: a verdict
// line, plus — only for CHANGED — the tree of non-Equal edits.
type consoleReporter struct{}

func (consoleReporter) Report(lock string, st status, edits []resultspecs.TestEdit) error {
	switch st {
	case statusUnchanged:
		p.Unchanged(lock)
	case statusNew:
		p.New(lock)
	case statusChanged:
		printEdits(edits)
		p.Changed(lock)
	}
	return nil
}

func printEdits(edits []resultspecs.TestEdit) {
	for _, edit := range edits {
		if edit.Action == resultspecs.Equal {
			continue
		}

		p.TestEdit(edit)
		for _, cmdedit := range edit.Commands {
			if cmdedit.Action == resultspecs.Equal {
				continue
			}

			p.CommandEdit(cmdedit)
			for _, chkedit := range cmdedit.Checks {
				if chkedit.Action == resultspecs.Equal {
					continue
				}

				p.CheckEdit(chkedit)
				for _, lineedit := range chkedit.Lines {
					p.LineEdit(lineedit)
				}
			}
		}
	}
}
