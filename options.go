package spellchecker

import (
	"bufio"
	"math"

	"github.com/agext/levenshtein"
)

// WithOpt set spellchecker options
func (s *Spellchecker) WithOpts(opts ...OptionFunc) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, o := range opts {
		if err := o(s); err != nil {
			return err
		}
	}

	return nil
}

// WithSplitter set splitter func for AddFrom() reader
func WithSplitter(f bufio.SplitFunc) OptionFunc {
	return func(s *Spellchecker) error {
		s.splitter = f
		return nil
	}
}

// WithMaxErrors sets maxErrors — the maximum allowed difference in bits
// between the "search word" and a "dictionary word".
// - deletion is a 1-bit change (proble → problem)
// - insertion is a 1-bit change (problemm → problem)
// - substitution is a 2-bit change (problam → problem)
// - transposition is a 0-bit change (problme → problem)
//
// It is not recommended to set this value greater than 2,
// as it can significantly affect performance.
func WithMaxErrors(maxErrors int) OptionFunc {
	return func(s *Spellchecker) error {
		s.maxErrors = maxErrors

		return nil
	}
}

// FilterFunc compares the source word with a candidate word.
// It returns the candidate's score and a boolean flag.
// If the flag is false, the candidate will be completely filtered out.
type FilterFunc func(src, candidate []rune, count uint) (float64, bool)

// WithFilterFunc set custom scoring function
func WithFilterFunc(f FilterFunc) OptionFunc {
	return func(s *Spellchecker) error {
		s.filterFunc = f
		return nil
	}
}

func defaultFilterFunc(maxErrors int) FilterFunc {
	return func(src, candidate []rune, count uint) (float64, bool) {
		distance, prefixLen, suffixLen := levenshtein.Calculate(src, candidate, 0, 1, 1, 1)
		if distance > maxErrors {
			return 0, false
		}

		mult := math.Log1p(float64(count)) * math.Pow(1.5, float64(prefixLen+suffixLen))

		return 1 / (1 + float64(distance*distance)) * mult, true
	}
}
