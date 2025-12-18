package spellchecker

import (
	"math"

	"github.com/agext/levenshtein"
)

const DefaultMaxErrors = 2

// FilterFunc compares the source word with a candidate word.
// It returns the candidate's score and a boolean flag.
// If the flag is false, the candidate will be completely filtered out.
type FilterFunc func(src, candidate []rune, count uint) (float64, bool)

type SearchOptionFunc func(opts *searchOptions)

// SuggestWithMaxErrors sets the maximum allowed difference in bits
// between the "search word" and a "dictionary word".
// - deletion is a 1-bit change (proble → problem)
// - insertion is a 1-bit change (problemm → problem)
// - substitution is a 2-bit change (problam → problem)
// - transposition is a 0-bit change (problme → problem)
//
// It is not recommended to set this value greater than 2,
// as it can significantly affect performance.
func SuggestWithMaxErrors(maxErrors int) SearchOptionFunc {
	return func(opts *searchOptions) {
		opts.maxErrors = maxErrors
	}
}

// SuggestWithFilterFunc set a FilterFunc
func SuggestWithFilterFunc(f FilterFunc) SearchOptionFunc {
	return func(opts *searchOptions) {
		opts.filterFunc = f
	}
}

type SuggestionResult struct {
	ExactMatch  bool // if true, the word is correct
	Suggestions []Match
}

// Suggest find top n suggestions for the word.
// Returns spellchecker scores along with words
func (s *Spellchecker) Suggest(word string, n int, opts ...SearchOptionFunc) SuggestionResult {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.has(word) {
		return SuggestionResult{ExactMatch: true}
	}

	searchOpts := defaultSearchOptions
	for _, o := range opts {
		o(&searchOpts)
	}

	return SuggestionResult{
		Suggestions: s.dict.find(word, n, searchOpts.maxErrors, searchOpts.filterFunc),
	}
}

type searchOptions struct {
	maxErrors  int
	filterFunc FilterFunc
}

var defaultSearchOptions = searchOptions{
	maxErrors:  DefaultMaxErrors,
	filterFunc: defaultFilterFunc(DefaultMaxErrors),
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
