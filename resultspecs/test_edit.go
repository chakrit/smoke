package resultspecs

import "github.com/chakrit/gendiff"

type TestEdit struct {
	Name     string
	Action   Action
	Commands []CommandEdit
}

type testDiff struct {
	oldSpecs []TestResultSpec
	newSpecs []TestResultSpec
}

var _ gendiff.Interface = testDiff{}

func (d testDiff) LeftLen() int        { return len(d.oldSpecs) }
func (d testDiff) RightLen() int       { return len(d.newSpecs) }
func (d testDiff) Equal(l, r int) bool { return d.oldSpecs[l].Name == d.newSpecs[r].Name }

func compareTests(oldspecs []TestResultSpec, newspecs []TestResultSpec) (edits []TestEdit, differs bool, err error) {
	diffs := gendiff.Make(testDiff{oldspecs, newspecs})

	for _, d := range diffs {
		switch d.Op {
		case gendiff.Delete:
			for lidx := d.Lstart; lidx < d.Lend; lidx++ {
				differs, edits = true, append(edits, TestEdit{
					Name:     oldspecs[lidx].Name,
					Action:   Removed,
					Commands: removedCommands(oldspecs[lidx].Commands),
				})
			}

		case gendiff.Insert:
			for ridx := d.Rstart; ridx < d.Rend; ridx++ {
				differs, edits = true, append(edits, TestEdit{
					Name:     newspecs[ridx].Name,
					Action:   Added,
					Commands: addedCommands(newspecs[ridx].Commands),
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
