package spellchecker

import (
	"io"
	"sync"
)

type Spellchecker struct {
	mtx sync.RWMutex

	dict *dictionary
}

func New(alphabet string) (*Spellchecker, error) {
	dict, err := newDictionary(alphabet)
	if err != nil {
		return nil, err
	}

	result := &Spellchecker{dict: dict}

	return result, nil
}

// AddFrom reads input, splits it with spellchecker splitter func and adds words to the dictionary
func (m *Spellchecker) AddFrom(opts *AddOptions, input io.Reader) error {
	if opts == nil {
		opts = defaultAddOptions
	}

	words := make([]string, 1000)
	i := 0
	for item := range readInput(input, opts.Splitter) {
		if item.err != nil {
			return item.err
		}

		if i == len(words) {
			m.Add(opts, words...)
			i = 0
		}
		words[i] = item.word
		i++
	}

	if i > 0 {
		m.Add(opts, words[:i]...)
	}

	return nil
}

// Add adds provided words to the dictionary with a custom weight
func (m *Spellchecker) Add(opts *AddOptions, words ...string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if opts == nil {
		opts = defaultAddOptions
	}

	for _, word := range words {
		if id := m.dict.id(word); id > 0 {
			m.dict.inc(id, opts.Weight)
			continue
		}

		m.dict.add(word, opts.Weight)
	}
}

// IsCorrect check if provided word is in the dictionary
func (s *Spellchecker) IsCorrect(word string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.dict.has(word)
}

func (s *Spellchecker) Fix(opts *SearchOptions, word string) (string, bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.has(word) {
		return word, true
	}

	opts = applyDefaults(opts)

	hits := s.dict.find(word, 1, opts.MaxErrors, opts.FilterFunc)
	if len(hits) == 0 {
		return word, false
	}

	return hits[0].Value, false
}

type SuggestionResult struct {
	ExactMatch  bool // if true, the word is correct
	Suggestions []Match
}

// Suggest find top n suggestions for the word.
// Returns spellchecker scores along with words
func (s *Spellchecker) Suggest(opts *SearchOptions, word string, n int) SuggestionResult {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.has(word) {
		return SuggestionResult{ExactMatch: true}
	}

	opts = applyDefaults(opts)

	return SuggestionResult{
		Suggestions: s.dict.find(word, n, opts.MaxErrors, opts.FilterFunc),
	}
}
