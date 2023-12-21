package resultspecs

import "github.com/chakrit/gendiff"

type CheckEdit struct {
	Name   string
	Action Action
	Lines  []LineEdit
}

type checkDiff struct {
	oldChks []CheckOutputSpec
	newChks []CheckOutputSpec
}

var _ gendiff.Interface = checkDiff{}

func (d checkDiff) LeftLen() int        { return len(d.oldChks) }
func (d checkDiff) RightLen() int       { return len(d.newChks) }
func (d checkDiff) Equal(l, r int) bool { return d.oldChks[l].Name == d.newChks[r].Name }

func addedChecks(checks []CheckOutputSpec) []CheckEdit {
	edits := make([]CheckEdit, 0, len(checks))
	for _, check := range checks {
		edits = append(edits, CheckEdit{
			Name:   check.Name,
			Action: Added,
			Lines:  addedLines(check.Data),
		})
	}
	return edits

}

func removedChecks(checks []CheckOutputSpec) []CheckEdit {
	edits := make([]CheckEdit, 0, len(checks))
	for _, check := range checks {
		edits = append(edits, CheckEdit{
			Name:   check.Name,
			Action: Removed,
			Lines:  removedLines(check.Data),
		})
	}
	return edits
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
