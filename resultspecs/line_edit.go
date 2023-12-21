package resultspecs

import (
	"github.com/chakrit/gendiff"
	"strconv"
)

type LineEdit struct {
	Line   string
	Action Action
}

type lineDiff struct {
	oldLines []string
	newLines []string
}

var _ gendiff.Interface = lineDiff{}

func (d lineDiff) LeftLen() int        { return len(d.oldLines) }
func (d lineDiff) RightLen() int       { return len(d.newLines) }
func (d lineDiff) Equal(l, r int) bool { return d.oldLines[l] == d.newLines[r] }

func addedLines(lines []string) []LineEdit {
	edits := make([]LineEdit, 0, len(lines))
	for _, line := range lines {
		edits = append(edits, LineEdit{
			Action: Added,
			Line:   line,
		})
	}
	return edits
}

func removedLines(lines []string) []LineEdit {
	edits := make([]LineEdit, 0, len(lines))
	for _, line := range lines {
		edits = append(edits, LineEdit{
			Action: Removed,
			Line:   line,
		})
	}
	return edits
}

func compareLines(oldlines []string, newlines []string) (edits []LineEdit, differs bool, err error) {
	diffs := gendiff.Make(lineDiff{oldlines, newlines})
	diffs = gendiff.Compact(diffs, 2)

	prev := gendiff.Diff{Op: -1}
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
			if prev.Op == gendiff.Match {
				// we have a gap (MMMMMMM -> MM...MM)
				// insert an indicator
				skips := d.Rstart - prev.Rend
				edits = append(edits, LineEdit{
					Action: Equal,
					Line:   " ... " + strconv.Itoa(skips) + " line(s) skipped ...",
				})
			}

			for lidx, ridx := d.Lstart, d.Rstart; lidx < d.Lend && ridx < d.Rend; lidx, ridx = lidx+1, ridx+1 {
				edits = append(edits, LineEdit{
					Line:   oldlines[lidx],
					Action: Equal,
				})
			}

		}

		prev = d
	}

	return
}
