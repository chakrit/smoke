package resultspecs

import "github.com/chakrit/gendiff"

type CommandEdit struct {
	Name   string
	Action Action
	Checks []CheckEdit
}

type commandDiff struct {
	oldCmds []CommandResultSpec
	newCmds []CommandResultSpec
}

var _ gendiff.Interface = commandDiff{}

func (d commandDiff) LeftLen() int        { return len(d.oldCmds) }
func (d commandDiff) RightLen() int       { return len(d.newCmds) }
func (d commandDiff) Equal(l, r int) bool { return d.oldCmds[l].Command == d.newCmds[r].Command }

func addedCommands(cmds []CommandResultSpec) []CommandEdit {
	edits := make([]CommandEdit, 0, len(cmds))
	for _, cmd := range cmds {
		edits = append(edits, CommandEdit{
			Name:   cmd.Command,
			Action: Added,
			Checks: addedChecks(cmd.Checks),
		})
	}
	return edits

}

func removedCommands(cmds []CommandResultSpec) []CommandEdit {
	edits := make([]CommandEdit, 0, len(cmds))
	for _, cmd := range cmds {
		edits = append(edits, CommandEdit{
			Name:   cmd.Command,
			Action: Removed,
			Checks: removedChecks(cmd.Checks),
		})
	}
	return edits
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
