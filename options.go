package spellchecker

import (
	"bufio"
	"math"
)

// WithSplitter set splitter func for AddFrom() reader
func WithSplitter(f bufio.SplitFunc) OptionFunc {
	return func(s *Spellchecker) error {
		s.splitter = f
		return nil
	}
}

// WithMaxErrors sets maxErrors — the maximum allowed difference in bits
// between the "search word" and a "dictionary word".
// For example, replacing a single character (problam => problem)
// is treated as a two-bit difference.
// It is not recommended to set a value greater than 2,
// as it can significantly impact performance.
func WithMaxErrors(maxErrors int) OptionFunc {
	return func(s *Spellchecker) error {
		s.maxErrors = maxErrors
		return nil
	}
}

type ScoreFunc = scoreFunc

// WithScoreFunc specify a function that will be used for scoring
func WithScoreFunc(f ScoreFunc) OptionFunc {
	return func(s *Spellchecker) error {
		s.dict.scoreFunc = f
		return nil
	}
}

var defaultScorefunc scoreFunc = func(src, candidate []rune, distance int, cnt uint) float64 {
	mult := math.Log1p(float64(cnt))
	// if first letters are the same, increase score
	if src[0] == candidate[0] {
		mult *= 1.5
		// if second letters are the same too, increase score even more
		if len(src) > 1 && len(candidate) > 1 && src[1] == candidate[1] {
			mult *= 1.5
		}
	}

	return 1 / (1 + float64(distance*distance)) * mult
}
