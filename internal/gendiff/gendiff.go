package gendiff

type Op int

const (
	noOp = Op(iota)
	Match
	Delete
	Insert
)

func (o Op) String() string {
	switch o {
	case noOp:
		return "noOp"
	case Match:
		return "match"
	case Delete:
		return "delete"
	case Insert:
		return "insert"
	default:
		panic("missing case in op.String() switch")
	}
}

type Interface interface {
	LeftLen() int
	RightLen() int
	Equal(l, r int) bool
}

type Diff struct {
	Op     Op
	Lstart int
	Lend   int
	Rstart int
	Rend   int
}

type cell = struct {
	op  Op  // detected operation
	lcs int // cumulative length of longest common substring
}

func Make(iface Interface) []Diff {
	var (
		llen, rlen = iface.LeftLen(), iface.RightLen()
		lidx, ridx = 0, 0
	)

	// table for dynamic programming an LCS solution
	// "left" is Y, "right" is X,
	// so index with table[y][x] or table[left str index][right str index]
	table := make([][]cell, llen+1, llen+1)
	for lidx = range table {
		table[lidx] = make([]cell, rlen+1, rlen+1)
	}

	// zeroes the empty string solution
	// (all inserts or all deletes)
	for lidx = range table { // empty "right" string, all were deleted
		table[lidx][0] = cell{Delete, 0}
	}
	for ridx = range table[0] { // empty "left" string, all were inserted
		table[0][ridx] = cell{Insert, 0}
	}

	// compute lcs solution (and obtain the diff as a sideeffect)
	for lidx = 1; lidx <= llen; lidx++ {
		for ridx = 1; ridx <= rlen; ridx++ {
			var (
				lcell  = table[lidx][ridx-1]   // neighbor towards the "left" string (x-1)
				rcell  = table[lidx-1][ridx]   // neighbor towards the "right" string (y-1)
				lrcell = table[lidx-1][ridx-1] // diagonal neighbor (x-1, y-1)
			)

			switch {
			case iface.Equal(lidx-1, ridx-1):
				// character match, extends the lcs counter
				table[lidx][ridx] = cell{op: Match, lcs: lrcell.lcs + 1}
			case lcell.lcs < rcell.lcs:
				// the "right" string has longer lcs which means we are sitting
				// on characters being deleted from the "left" string
				table[lidx][ridx] = cell{op: Delete, lcs: rcell.lcs}
			case lcell.lcs >= rcell.lcs:
				// the "left" string has longer lcs which means we are sitting
				// on an extra characters from the "right" string
				table[lidx][ridx] = cell{op: Insert, lcs: lcell.lcs}
			}
		}
	}

	// reconstruct solution backwards
	var (
		diffs    []Diff
		lastcell = table[llen][rlen]
		lastdiff = Diff{lastcell.op, llen, llen, rlen, rlen}
	)

	record := func(op Op, lidx, ridx int) {
		lastdiff.Lstart = lidx
		lastdiff.Rstart = ridx
		if op != lastdiff.Op {
			diffs = append(diffs, lastdiff)
			lastdiff.Op = op
			lastdiff.Lend = lastdiff.Lstart
			lastdiff.Rend = lastdiff.Rstart
		}
	}

	lidx, ridx = llen, rlen
	for lidx > 0 || ridx > 0 {
		cell := table[lidx][ridx]
		record(cell.op, lidx, ridx)

		switch cell.op {
		case Match:
			lidx, ridx = lidx-1, ridx-1
		case Delete:
			lidx, ridx = lidx-1, ridx
		case Insert:
			lidx, ridx = lidx, ridx-1
		default:
			panic("DP table construction error, please file a bug report.")
		}
	}

	record(noOp, 0, 0) // eof signal to emit the last diff

	// since we construct solution backwards, we need to reverse it
	revdiffs := make([]Diff, len(diffs), len(diffs))
	for idx := range diffs {
		revdiffs[len(diffs)-idx-1] = diffs[idx]
	}
	return revdiffs
}
