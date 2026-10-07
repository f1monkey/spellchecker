package levenshtein

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Levenshtein(t *testing.T) {
	t.Parallel()

	t.Run("must calculate known distances", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			a, b           string
			maxDist        int
			dist, pre, suf int
		}{
			{a: "", b: "", maxDist: 2, dist: 0},
			{a: "abc", b: "abc", maxDist: 2, dist: 0, pre: 3},
			{a: "", b: "ab", maxDist: 2, dist: 2},
			{a: "ab", b: "", maxDist: 2, dist: 2},
			{a: "kitten", b: "sitting", maxDist: 5, dist: 3, pre: 0, suf: 0},
			{a: "kitten", b: "sitting", maxDist: 2, dist: 3},
			{a: "arang", b: "range", maxDist: 2, dist: 2},
			{a: "arang", b: "orange", maxDist: 2, dist: 2, suf: 0},
			{a: "arang", b: "green", maxDist: 2, dist: 3},
			{a: "problam", b: "problem", maxDist: 2, dist: 1, pre: 5, suf: 1},
			{a: "oragne", b: "orange", maxDist: 2, dist: 2, pre: 3, suf: 1},
			{a: "proble", b: "problem", maxDist: 2, dist: 1, pre: 6},
			{a: "abc", b: "abcdef", maxDist: 2, dist: 3, pre: 3},
			{a: "молоко", b: "малако", maxDist: 2, dist: 2, pre: 1, suf: 2},
		}

		for _, c := range cases {
			dist, pre, suf := Levenshtein([]rune(c.a), []rune(c.b), c.maxDist)
			require.Equal(t, c.dist, dist, "%q → %q, max %d", c.a, c.b, c.maxDist)
			require.Equal(t, c.pre, pre, "prefix %q → %q", c.a, c.b)
			require.Equal(t, c.suf, suf, "suffix %q → %q", c.a, c.b)
		}
	})

	t.Run("must match full DP on random strings", func(t *testing.T) {
		t.Parallel()

		rnd := rand.New(rand.NewPCG(1, 2)) //nolint:gosec

		for range 20000 {
			alphabet := []rune("abc")
			if rnd.IntN(2) == 0 {
				alphabet = []rune("abcdefghijklmnopqrstuvwxyz")
			}

			maxLen := 10
			if rnd.IntN(20) == 0 {
				maxLen = 2 * stackRowSize
			}

			a := randomRunes(rnd, alphabet, rnd.IntN(maxLen+1))
			b := mutateRunes(rnd, alphabet, a, rnd.IntN(4))
			maxDist := rnd.IntN(5)

			expected := min(referenceLevenshtein(a, b), maxDist+1)
			dist, _, _ := Levenshtein(a, b, maxDist)
			require.Equal(t, expected, dist, "%q → %q, max %d", string(a), string(b), maxDist)
		}
	})
}

//nolint:paralleltest // testing.AllocsPerRun panics in parallel tests
func Test_levenshtein_NoAllocs(t *testing.T) {
	a, b := []rune("problam"), []rune("problems")

	allocs := testing.AllocsPerRun(100, func() {
		Levenshtein(a, b, 2)
	})
	require.Zero(t, allocs)
}

func referenceLevenshtein(a, b []rune) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		cur[0] = i

		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			cur[j] = min(prev[j-1]+cost, prev[j]+1, cur[j-1]+1)
		}

		prev, cur = cur, prev
	}

	return prev[len(b)]
}

func randomRunes(rnd *rand.Rand, alphabet []rune, n int) []rune {
	result := make([]rune, n)
	for i := range result {
		result[i] = alphabet[rnd.IntN(len(alphabet))]
	}

	return result
}

// mutateRunes applies the given number of random insertions, deletions and substitutions.
func mutateRunes(rnd *rand.Rand, alphabet []rune, src []rune, edits int) []rune {
	result := append([]rune(nil), src...)

	for range edits {
		letter := alphabet[rnd.IntN(len(alphabet))]

		switch op := rnd.IntN(3); {
		case op == 0 || len(result) == 0:
			pos := rnd.IntN(len(result) + 1)
			result = append(result[:pos], append([]rune{letter}, result[pos:]...)...)
		case op == 1:
			pos := rnd.IntN(len(result))
			result = append(result[:pos], result[pos+1:]...)
		default:
			result[rnd.IntN(len(result))] = letter
		}
	}

	return result
}

func Benchmark_levenshtein(b *testing.B) {
	src, candidate := []rune("problam"), []rune("problems")

	for b.Loop() {
		Levenshtein(src, candidate, 2)
	}
}
