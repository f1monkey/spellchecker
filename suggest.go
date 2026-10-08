package spellchecker

import (
	"math"

	"github.com/f1monkey/spellchecker/v4/internal/dictionary"
	"github.com/f1monkey/spellchecker/v4/internal/levenshtein"
)

const DefaultMaxErrors = 2

// FilterFunc compares the source word with a candidate word.
// It returns the candidate's score and a boolean flag.
// If the flag is false, the candidate will be completely filtered out.
type FilterFunc = dictionary.FilterFunc

type OptionFunc func(opts *searchOptions)

// WithMaxErrors sets the maximum allowed difference in bits
// between the "search word" and a "dictionary word".
// - deletion is a 1-bit change (proble → problem)
// - insertion is a 1-bit change (problemm → problem)
// - substitution is a 2-bit change (problam → problem)
// - transposition is a 0-bit change (problme → problem)
//
// It is not recommended to set this value greater than 2,
// as it can significantly affect performance.
func WithMaxErrors(maxErrors int) OptionFunc {
	return func(opts *searchOptions) {
		opts.maxErrors = maxErrors
	}
}

// WithFilterFunc set a FilterFunc
func WithFilterFunc(f FilterFunc) OptionFunc {
	return func(opts *searchOptions) {
		opts.filterFunc = f
	}
}

type Match = dictionary.Match

type SuggestionResult struct {
	ExactMatch  bool // if true, the word is correct
	Suggestions []Match
}

// Suggest find top n suggestions for the word.
// Returns spellchecker scores along with words
func (s *Spellchecker) Suggest(word string, n int, opts ...OptionFunc) SuggestionResult {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.Has(word) {
		return SuggestionResult{ExactMatch: true}
	}

	searchOpts := searchOptions{maxErrors: DefaultMaxErrors, filterFunc: defaultFilterFunc}
	for _, o := range opts {
		o(&searchOpts)
	}

	if searchOpts.filterFunc == nil {
		searchOpts.filterFunc = defaultFilterFunc
	}

	return SuggestionResult{
		Suggestions: s.dict.Find(word, n, searchOpts.maxErrors, searchOpts.filterFunc),
	}
}

type searchOptions struct {
	maxErrors  int
	filterFunc FilterFunc
}

var defaultFilterFunc FilterFunc = func(src, candidate []rune, count uint, maxErrors int) (float64, bool) {
	const prefixCoefficitent = 1.5

	if math.Abs(float64(len(src)-len(candidate))) > float64(maxErrors) {
		return 0, false
	}

	distance, prefixLen, suffixLen := levenshtein.Levenshtein(src, candidate, maxErrors)
	if distance > maxErrors {
		return 0, false
	}

	mult := math.Log1p(float64(count)) * math.Pow(prefixCoefficitent, float64(prefixLen+suffixLen))

	return 1 / (1 + float64(distance*distance)) * mult, true
}
