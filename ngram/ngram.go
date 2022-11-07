package ngram

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type NGram struct {
	Value    string
	Position int
}

// Make generates ngrams with len=1..max. Panic if max < 1
func MakeAll(word string, max int) []NGram {
	if max < 1 {
		panic("max must be >= 1")
	}
	if word == "" {
		return nil
	}

	wordLen := utf8.RuneCountInString(word)
	maxN := min(wordLen, max)

	result := make([]NGram, 0)
	for n := 1; n <= maxN; n++ {
		result = append(result, Make(word, n)...)
	}

	return result
}

// Make generates ngrams with len=n
func Make(word string, n int) []NGram {
	if n < 1 {
		panic(fmt.Errorf("unable to make trigrams with n=%d", n))
	}

	runes := []rune(word)
	if n > len(runes) {
		return nil
	}

	if n == len(runes) {
		return []NGram{{Value: word}}
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

	result := make([]NGram, cnt)
	for i, b := range builders {
		result[i] = NGram{
			Value:    b.String(),
			Position: i,
		}
	}

	return result
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
