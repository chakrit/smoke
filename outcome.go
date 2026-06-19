package main

import (
	"errors"

	"github.com/chakrit/smoke/internal/p"
)

// A fatal ends a multi-spec run early (fail-fast). Two kinds, distinguished by
// type so the single exit authority in main maps each to its frozen code:
//
//   - dataError → malformed input (a spec or lock read but unparseable/invalid),
//     exit 65 (EX_DATAERR).
//   - any other error → operational trouble, exit 2.
//
// reported additionally marks a fatal that was already surfaced live (the run
// hooks print as they go), so main doesn't print it a second time.
type dataError struct{ err error }

func (e *dataError) Error() string { return e.err.Error() }
func (e *dataError) Unwrap() error { return e.err }

type reported struct{ err error }

func (e *reported) Error() string { return e.err.Error() }
func (e *reported) Unwrap() error { return e.err }

func dataErr(err error) error     { return &dataError{err} }
func reportedErr(err error) error { return &reported{err} }

// exitCode maps a fatal to its frozen exit code — malformed input → 65, else
// operational → 2. See docs/spec/exit-codes.md.
func exitCode(err error) int {
	var de *dataError
	if errors.As(err, &de) {
		return p.ExitDataErr
	}
	return p.ExitTrouble
}

// wasReported reports whether the fatal was already printed live, so the exit
// authority skips re-printing it.
func wasReported(err error) bool {
	var r *reported
	return errors.As(err, &r)
}
