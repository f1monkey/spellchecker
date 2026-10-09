package spellchecker

// stackRowSize is the max word length whose row is allocated on the stack.
const stackRowSize = 64

// levenshtein returns the distance between a and b, capped at maxDist+1,
// and the lengths of their common prefix and suffix.
func levenshtein(a, b []rune, maxDist int) (dist, prefixLen, suffixLen int) {
	for prefixLen < len(a) && prefixLen < len(b) && a[prefixLen] == b[prefixLen] {
		prefixLen++
	}

	a, b = a[prefixLen:], b[prefixLen:]

	for suffixLen < len(a) && suffixLen < len(b) && a[len(a)-1-suffixLen] == b[len(b)-1-suffixLen] {
		suffixLen++
	}

	a, b = a[:len(a)-suffixLen], b[:len(b)-suffixLen]

	return boundedDistance(a, b, maxDist), prefixLen, suffixLen
}

func boundedDistance(a, b []rune, maxDist int) int {
	if maxDist < 0 {
		maxDist = 0
	}

	limit := maxDist + 1

	n, m := len(a), len(b)
	if diff := n - m; diff > maxDist || -diff > maxDist {
		return limit
	}

	if n == 0 {
		return m
	}

	if m == 0 {
		return n
	}

	var stack [stackRowSize + 1]int

	row := stack[:]
	if m+1 > len(stack) {
		row = make([]int, m+1)
	}

	row = row[:m+1]
	for j := range row {
		row[j] = min(j, limit)
	}

	for i := 1; i <= n; i++ {
		lo := max(1, i-maxDist)
		hi := min(m, i+maxDist)

		diag := row[lo-1]
		if lo == 1 {
			row[0] = min(i, limit)
		} else {
			row[lo-1] = limit
		}

		rowMin := row[lo-1]

		for j := lo; j <= hi; j++ {
			up := row[j]

			v := diag
			if a[i-1] != b[j-1] {
				v++
			}

			v = min(v, up+1, row[j-1]+1, limit)

			diag = up
			row[j] = v
			rowMin = min(rowMin, v)
		}

		if rowMin >= limit {
			return limit
		}
	}

	return row[m]
}
