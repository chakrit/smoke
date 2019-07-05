package resultspecs

import (
	"strings"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/xerrors"
)

type (
	Action int

	TestEdit = struct {
		Name     string
		Action   Action
		Commands []CommandEdit
	}

	CommandEdit struct {
		Name   string
		Action Action
		Checks []CheckEdit
	}

	CheckEdit struct {
		Name   string
		Action Action
		Diffs  []dmp.Diff
	}
)

const (
	Equal = Action(iota)
	Added
	Removed
	InnerChanges
)

func Compare(oldspecs []TestResultSpec, newspecs []TestResultSpec) (edits []TestEdit, differs bool, err error) {
	emit := func(edit TestEdit) {
		differs = differs || edit.Action != Equal
		edits = append(edits, edit)
	}

	oldidx, newidx := 0, 0
	for oldidx < len(oldspecs) && newidx < len(newspecs) {
		oldspec, newspec := oldspecs[oldidx], newspecs[newidx]

		// delete before inserts
		if oldspec.Name != newspec.Name {
			emit(TestEdit{oldspec.Name, Removed, nil})
			oldidx += 1
			continue
		}

		if cmdedits, differs_, err := compareCommands(oldspec.Commands, newspec.Commands); err != nil {
			return nil, true, xerrors.Errorf("compare: %w", err)
		} else if !differs_ {
			emit(TestEdit{oldspec.Name, Equal, nil})
		} else {
			emit(TestEdit{oldspec.Name, InnerChanges, cmdedits})
		}

		oldidx += 1
		newidx += 1
	}

	// oldspecs or newspecs have EOF, these are either all-deletes or all-inserts
	// we search for delete before inserts
	for ; oldidx < len(oldspecs); oldidx++ {
		emit(TestEdit{oldspecs[oldidx].Name, Removed, nil})
	}
	for ; newidx < len(newspecs); newidx++ {
		emit(TestEdit{newspecs[newidx].Name, Added, nil})
	}
	return
}

func compareCommands(oldcmds []CommandResultSpec, newcmds []CommandResultSpec) (edits []CommandEdit, differs bool, err error) {
	emit := func(edit CommandEdit) {
		differs = differs || edit.Action != Equal
		edits = append(edits, edit)
	}

	oldidx, newidx := 0, 0
	for oldidx < len(oldcmds) && newidx < len(newcmds) {
		oldcmd, newcmd := oldcmds[oldidx], newcmds[newidx]

		// delete before inserts
		if oldcmd.Command != newcmd.Command {
			emit(CommandEdit{oldcmd.Command, Removed, nil})
			oldidx += 1
			continue
		}

		if chkedits, differs_, err := compareChecks(oldcmd.Checks, newcmd.Checks); err != nil {
			return nil, true, xerrors.Errorf("compare: %w", err)
		} else if !differs_ {
			emit(CommandEdit{oldcmd.Command, Equal, nil})
		} else {
			emit(CommandEdit{oldcmd.Command, InnerChanges, chkedits})
		}

		oldidx += 1
		newidx += 1
	}

	// delete before inserts
	for ; oldidx < len(oldcmds); oldidx++ {
		emit(CommandEdit{oldcmds[oldidx].Command, Removed, nil})
	}
	for ; newidx < len(newcmds); newidx++ {
		emit(CommandEdit{newcmds[newidx].Command, Added, nil})
	}
	return
}

func compareChecks(oldchks []CheckOutputSpec, newchks []CheckOutputSpec) (edits []CheckEdit, differs bool, err error) {
	emit := func(edit CheckEdit) {
		differs = differs || edit.Action != Equal
		edits = append(edits, edit)
	}

	oldidx, newidx := 0, 0
	for oldidx < len(oldchks) && newidx < len(newchks) {
		oldchk, newchk := oldchks[oldidx], newchks[newidx]

		// delete before inserts
		if oldchk.Name != newchk.Name {
			emit(CheckEdit{oldchk.Name, Removed, nil})
			oldidx += 1
			continue
		}

		oldoutput := strings.Join(oldchk.Data, "\n")
		newoutput := strings.Join(newchk.Data, "\n")

		diffs := dmp.New().DiffMain(oldoutput, newoutput, false)
		if len(diffs) == 0 {
			emit(CheckEdit{oldchk.Name, Equal, nil})
		} else {
			allequals := true
			for _, diff := range diffs {
				if diff.Type != dmp.DiffEqual {
					allequals = false
					break
				}
			}
			if !allequals {
				emit(CheckEdit{oldchk.Name, InnerChanges, diffs})
			} else {
				emit(CheckEdit{oldchk.Name, Equal, nil})
			}
		}

		oldidx += 1
		newidx += 1
	}

	// delete before inserts
	for ; oldidx < len(oldchks); oldidx++ {
		emit(CheckEdit{oldchks[oldidx].Name, Removed, nil})
	}
	for ; newidx < len(newchks); newidx++ {
		emit(CheckEdit{newchks[newidx].Name, Added, nil})
	}
	return
}
