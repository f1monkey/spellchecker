package spellchecker

import (
	"encoding/gob"
	"io"
	"math"
	"sync"
)

const defaultMaxErrors = 2

type dict interface {
	ID(word string) uint32
	Has(word string) bool
	Inc(id uint32, n uint)
	Add(word string, n uint) uint32
	Find(word string, n int, maxErrors int, fn ScoringFunc) []Match
}

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

// WithScoringFunc set a ScoringFunc
func WithScoringFunc(f ScoringFunc) OptionFunc {
	return func(opts *searchOptions) {
		opts.scoringFunc = f
	}
}

type Spellchecker struct {
	mtx sync.RWMutex

	dict dict
}

// New creates a spellchecker with the given tokenizer and alphabets.
// Alphabets are the allowed characters for indexing and lookup; other
// characters are ignored. Pass one or more constants such as EN, RU, Numbers,
// or any custom string.
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

// Add adds each argument as a whole dictionary word. Strings are not split;
// use AddPhrases to tokenize text.
func (s *Spellchecker) Add(words ...string) {
	s.AddWeight(1, words...)
}

// AddWeight is like Add, but each word is stored with the given frequency
// weight. If the word is already in the dictionary, the weight is added to
// its count. Higher weight ranks the word higher in suggestions.
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

type spellcheckerData struct {
	Dict *dictionary
}

// Save encodes spellchecker data and writes it to the provided writer
func (m *Spellchecker) Save(w io.Writer) error {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	//nolint:forcetypeassert
	data := spellcheckerData{
		Dict: m.dict.(*dictionary),
	}

	return gob.NewEncoder(w).Encode(data)
}

type Suggestion = Match

type SuggestionResult struct {
	ExactMatch  bool // if true, the word is correct
	Suggestions []Suggestion
}

// Suggest find top n suggestions for the word.
// Returns spellchecker scores along with words
func (s *Spellchecker) Suggest(word string, n int, opts ...OptionFunc) SuggestionResult {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.Has(word) {
		return SuggestionResult{ExactMatch: true}
	}

	searchOpts := searchOptions{maxErrors: defaultMaxErrors, scoringFunc: defaultScoringFunc}
	for _, o := range opts {
		o(&searchOpts)
	}

	if searchOpts.scoringFunc == nil {
		searchOpts.scoringFunc = defaultScoringFunc
	}

	return SuggestionResult{
		Suggestions: s.dict.Find(word, n, searchOpts.maxErrors, searchOpts.scoringFunc),
	}
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

// ScoringFunc compares the source word with a candidate word.
// It returns the candidate's score and a boolean flag.
// If the flag is false, the candidate will be completely filtered out.
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
