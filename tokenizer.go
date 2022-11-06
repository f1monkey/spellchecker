package spellchecker

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

const maxNGram = 3

type Term struct {
	Value    string
	Position int
}

func makeTerms(word string) []Term {
	maxN := int(math.Min(maxNGram, float64(utf8.RuneCountInString(word))))
	if maxN == 0 {
		return nil
	}

	result := make([]Term, 0)

	for n := 1; n <= maxN; n++ {
		for i, ngram := range makeNGrams(word, n) {
			result = append(result, Term{
				Value:    ngram,
				Position: i,
			})
		}
	}

	return result
}

func makeNGrams(word string, n int) []string {
	if n < 1 {
		panic(fmt.Errorf("unable to make trigrams with n=%d", n))
	}

	runes := []rune(word)
	if n > len(runes) {
		return nil
	}

	if n == len(runes) {
		return []string{word}
	}

	cnt := len(runes) - n + 1
	builders := make([]strings.Builder, cnt)

	for i, r := range []rune(word) {
		for j := i - n + 1; j <= i; j++ {
			if j < 0 {
				continue
			}
			if j >= cnt {
				break
			}

			builders[j].WriteRune(r)
		}
	}

	result := make([]string, cnt)
	for i, b := range builders {
		result[i] = b.String()
	}

	return result
}
