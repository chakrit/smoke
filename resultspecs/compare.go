package resultspecs

import (
	"github.com/chakrit/smoke/internal/gendiff"
)

const (
	Equal = Action(iota)
	Added
	Removed
	InnerChanges
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
		Lines  []LineEdit
	}
	LineEdit struct {
		Line   string
		Action Action
	}

	testDiff struct {
		oldSpecs []TestResultSpec
		newSpecs []TestResultSpec
	}
	commandDiff struct {
		oldCmds []CommandResultSpec
		newCmds []CommandResultSpec
	}
	checkDiff struct {
		oldChks []CheckOutputSpec
		newChks []CheckOutputSpec
	}
	lineDiff struct {
		oldLines []string
		newLines []string
	}
)

var (
	_ gendiff.Interface = testDiff{}
	_ gendiff.Interface = commandDiff{}
	_ gendiff.Interface = checkDiff{}
	_ gendiff.Interface = lineDiff{}
)

func (d testDiff) LeftLen() int           { return len(d.oldSpecs) }
func (d testDiff) RightLen() int          { return len(d.newSpecs) }
func (d testDiff) Equal(l, r int) bool    { return d.oldSpecs[l].Name == d.newSpecs[r].Name }
func (d commandDiff) LeftLen() int        { return len(d.oldCmds) }
func (d commandDiff) RightLen() int       { return len(d.newCmds) }
func (d commandDiff) Equal(l, r int) bool { return d.oldCmds[l].Command == d.newCmds[r].Command }
func (d checkDiff) LeftLen() int          { return len(d.oldChks) }
func (d checkDiff) RightLen() int         { return len(d.newChks) }
func (d checkDiff) Equal(l, r int) bool   { return d.oldChks[l].Name == d.newChks[r].Name }
func (d lineDiff) LeftLen() int           { return len(d.oldLines) }
func (d lineDiff) RightLen() int          { return len(d.newLines) }
func (d lineDiff) Equal(l, r int) bool    { return d.oldLines[l] == d.newLines[r] }

func Compare(oldspecs []TestResultSpec, newspecs []TestResultSpec) (edits []TestEdit, differs bool, err error) {
	diffs := gendiff.Make(testDiff{oldspecs, newspecs})

	for _, d := range diffs {
		switch d.Op {
		case gendiff.Delete:
			for lidx := d.Lstart; lidx < d.Lend; lidx++ {
				differs, edits = true, append(edits, TestEdit{
					Name:   oldspecs[lidx].Name,
					Action: Removed,
				})
			}

		case gendiff.Insert:
			for ridx := d.Rstart; ridx < d.Rend; ridx++ {
				differs, edits = true, append(edits, TestEdit{
					Name:   newspecs[ridx].Name,
					Action: Added,
				})
			}

		case gendiff.Match:
			for lidx, ridx := d.Lstart, d.Rstart; lidx < d.Lend && ridx < d.Rend; lidx, ridx = lidx+1, ridx+1 {
				oldspec, newspec := oldspecs[lidx], newspecs[ridx]
				if cmdedits, cmddiffers, err := compareCmds(oldspec.Commands, newspec.Commands); err != nil {
					return nil, true, err
				} else if cmddiffers {
					differs, edits = true, append(edits, TestEdit{
						Name:     oldspec.Name,
						Action:   InnerChanges,
						Commands: cmdedits,
					})
				} else {
					edits = append(edits, TestEdit{
						Name:   newspec.Name,
						Action: Equal,
					})
				}
			}
		}
	}

	return
}

func compareCmds(oldcmds []CommandResultSpec, newcmds []CommandResultSpec) (edits []CommandEdit, differs bool, err error) {
	diffs := gendiff.Make(commandDiff{oldcmds, newcmds})

	for _, d := range diffs {
		switch d.Op {
		case gendiff.Delete:
			for lidx := d.Lstart; lidx < d.Lend; lidx++ {
				differs, edits = true, append(edits, CommandEdit{
					Name:   oldcmds[lidx].Command,
					Action: Removed,
				})
			}

		case gendiff.Insert:
			for ridx := d.Rstart; ridx < d.Rend; ridx++ {
				differs, edits = true, append(edits, CommandEdit{
					Name:   newcmds[ridx].Command,
					Action: Added,
				})
			}

		case gendiff.Match:
			for lidx, ridx := d.Lstart, d.Rstart; lidx < d.Lend && ridx < d.Rend; lidx, ridx = lidx+1, ridx+1 {
				oldcmd, newcmd := oldcmds[lidx], newcmds[ridx]
				if chkedits, chkdiffers, err := compareChecks(oldcmd.Checks, newcmd.Checks); err != nil {
					return nil, true, err
				} else if chkdiffers {
					differs, edits = true, append(edits, CommandEdit{
						Name:   oldcmd.Command,
						Action: InnerChanges,
						Checks: chkedits,
					})
				} else {
					edits = append(edits, CommandEdit{
						Name:   newcmd.Command,
						Action: Equal,
					})
				}
			}

		}
	}

	return
}

func compareChecks(oldchks []CheckOutputSpec, newchks []CheckOutputSpec) (edits []CheckEdit, differs bool, err error) {
	diffs := gendiff.Make(checkDiff{oldchks, newchks})

	for _, d := range diffs {
		switch d.Op {
		case gendiff.Delete:
			for lidx := d.Lstart; lidx < d.Lend; lidx++ {
				differs, edits = true, append(edits, CheckEdit{
					Name:   oldchks[lidx].Name,
					Action: Removed,
				})
			}

		case gendiff.Insert:
			for ridx := d.Rstart; ridx < d.Rend; ridx++ {
				differs, edits = true, append(edits, CheckEdit{
					Name:   newchks[ridx].Name,
					Action: Added,
				})
			}

		case gendiff.Match:
			for lidx, ridx := d.Lstart, d.Rstart; lidx < d.Lend && ridx < d.Rend; lidx, ridx = lidx+1, ridx+1 {
				oldchk, newchk := oldchks[lidx], newchks[ridx]
				if lineedits, linediffers, err := compareLines(oldchk.Data, newchk.Data); err != nil {
					return nil, true, err
				} else if linediffers {
					differs, edits = true, append(edits, CheckEdit{
						Name:   oldchk.Name,
						Action: InnerChanges,
						Lines:  lineedits,
					})
				} else {
					edits = append(edits, CheckEdit{
						Name:   newchk.Name,
						Action: Equal,
					})
				}
			}

		}
	}

	return
}

func compareLines(oldlines []string, newlines []string) (edits []LineEdit, differs bool, err error) {
	diffs := gendiff.Make(lineDiff{oldlines, newlines})

	for _, d := range diffs {
		switch d.Op {
		case gendiff.Delete:
			for lidx := d.Lstart; lidx < d.Lend; lidx++ {
				differs, edits = true, append(edits, LineEdit{
					Line:   oldlines[lidx],
					Action: Removed,
				})
			}

		case gendiff.Insert:
			for ridx := d.Rstart; ridx < d.Rend; ridx++ {
				differs, edits = true, append(edits, LineEdit{
					Line:   newlines[ridx],
					Action: Added,
				})
			}

		case gendiff.Match:
			for lidx, ridx := d.Lstart, d.Rstart; lidx < d.Lend && ridx < d.Rend; lidx, ridx = lidx+1, ridx+1 {
				edits = append(edits, LineEdit{
					Line:   oldlines[lidx],
					Action: Equal,
				})
			}

		}
	}

	return
}
