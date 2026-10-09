package spellchecker

import (
	"encoding/gob"
	"io"
	"math"
	"sync"
)

const defaultMaxErrors = 2

// Mistake is a set of mistakes.
type Mistake int

// NoMistake means the segment is correct.
const NoMistake Mistake = 0

const (
	// MistakeTypo is a misspelled word.
	MistakeTypo Mistake = 1 << iota
	// MistakeUnknownWord is a word that is not in the dictionary and has no suggestions.
	MistakeUnknownWord
	// MistakeLayout is a word typed in a wrong keyboard layout.
	MistakeLayout
)

// Has reports whether m contains any of the mistakes in x.
func (m Mistake) Has(x Mistake) bool { return m&x != 0 }

type OptionFunc func(opts *searchOptions)

// WithMaxErrors sets the max allowed edit distance to a dictionary word.
// Values greater than 2 significantly affect performance.
func WithMaxErrors(maxErrors int) OptionFunc {
	return func(opts *searchOptions) {
		opts.maxErrors = maxErrors
	}
}

// WithScoringFunc set a ScoringFunc
func WithScoringFunc(f ScoringFunc) OptionFunc {
	return func(opts *searchOptions) {
		opts.scoringFunc = f
	}
}

type Spellchecker struct {
	mtx sync.RWMutex

	dict *dictionary
}

// New creates a spellchecker for the given alphabets. Other characters are ignored.
func New(alphabets ...Alphabet) (*Spellchecker, error) {
	dict, err := newDictionary(alphabets...)
	if err != nil {
		return nil, err
	}

	result := &Spellchecker{
		dict: dict,
	}

	return result, nil
}

// IsCorrect check if provided word is in the dictionary
func (s *Spellchecker) IsCorrect(word string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.dict.Has(word)
}

// Add adds each argument as a whole word.
func (s *Spellchecker) Add(words ...string) {
	s.AddWeight(1, words...)
}

// AddWeight adds words with the given weight. Higher weight ranks a word higher.
func (s *Spellchecker) AddWeight(weight uint, words ...string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, word := range words {
		if word == "" {
			continue
		}

		if id := s.dict.ID(word); id > 0 {
			s.dict.Inc(id, weight)

			continue
		}

		s.dict.Add(word, weight)
	}
}

// SuggestionResult is the result of Suggest.
type SuggestionResult struct {
	// Mistakes are mistakes found in the word.
	Mistakes Mistake
	// Suggestions are fix candidates, best first.
	Suggestions []Suggestion
}

// Suggest returns up to n fix suggestions for the word.
func (s *Spellchecker) Suggest(word string, n int, opts ...OptionFunc) SuggestionResult {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.Has(word) {
		return SuggestionResult{}
	}

	searchOpts := searchOptions{maxErrors: defaultMaxErrors, scoringFunc: defaultScoringFunc}
	for _, o := range opts {
		o(&searchOpts)
	}

	if searchOpts.scoringFunc == nil {
		searchOpts.scoringFunc = defaultScoringFunc
	}

	matches := s.dict.Find(word, n, searchOpts.maxErrors, searchOpts.scoringFunc)
	if len(matches) == 0 {
		return SuggestionResult{Mistakes: MistakeUnknownWord}
	}

	return SuggestionResult{
		Mistakes:    MistakeTypo,
		Suggestions: matches,
	}
}

type spellcheckerData struct {
	Dict *dictionary
}

// Save encodes spellchecker data and writes it to the provided writer
func (m *Spellchecker) Save(w io.Writer) error {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	//nolint:forcetypeassert
	data := spellcheckerData{
		Dict: m.dict,
	}

	return gob.NewEncoder(w).Encode(data)
}

// Load reads spellchecker data from the provided reader and decodes it.
func Load(reader io.Reader) (*Spellchecker, error) {
	data := spellcheckerData{}

	err := gob.NewDecoder(reader).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &Spellchecker{
		dict: data.Dict,
	}, nil
}

type searchOptions struct {
	maxErrors   int
	scoringFunc ScoringFunc
}

// ScoringFunc scores a candidate for the source word. false filters the candidate out.
type ScoringFunc func(src, candidate []rune, count uint, maxErrors int) (float64, bool)

var defaultScoringFunc ScoringFunc = func(src, candidate []rune, count uint, maxErrors int) (float64, bool) {
	const prefixCoefficitent = 1.5

	distance, prefixLen, suffixLen := levenshtein(src, candidate, maxErrors)
	if distance > maxErrors {
		return 0, false
	}

	mult := math.Log1p(float64(count)) * math.Pow(prefixCoefficitent, float64(prefixLen+suffixLen))

	return 1 / (1 + float64(distance*distance)) * mult, true
}
