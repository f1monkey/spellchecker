package spellchecker

import (
	"bufio"
	"fmt"
	"io"
	"sync"
)

const DefaultMaxErrors = 2

// OptionFunc option setter
type OptionFunc func(s *Spellchecker) error

type Spellchecker struct {
	mtx sync.RWMutex

	dict       *dictionary
	splitter   bufio.SplitFunc
	filterFunc FilterFunc
	scoreFunc  ScoreFunc
	maxErrors  int
}

func New(alphabet string, opts ...OptionFunc) (*Spellchecker, error) {
	result := &Spellchecker{
		maxErrors:  DefaultMaxErrors,
		filterFunc: defaultFilterFunc(DefaultMaxErrors),
	}

	for _, o := range opts {
		if err := o(result); err != nil {
			return nil, err
		}
	}

	if result.scoreFunc != nil {
		result.filterFunc = wrapScoreFunc(result.scoreFunc, result.maxErrors)
	}

	dict, err := newDictionary(alphabet, result.filterFunc, result.maxErrors)
	if err != nil {
		return nil, err
	}

	result.dict = dict

	return result, nil
}

// AddFrom reads input, splits it with spellchecker splitter func and adds words to the dictionary
func (m *Spellchecker) AddFrom(input io.Reader) error {
	words := make([]string, 1000)
	i := 0
	for item := range readInput(input, m.splitter) {
		if item.err != nil {
			return item.err
		}

		if i == len(words) {
			m.Add(words...)
			i = 0
		}
		words[i] = item.word
		i++
	}

	if i > 0 {
		m.Add(words[:i]...)
	}

	return nil
}

// Add adds provided words to the dictionary
func (m *Spellchecker) Add(words ...string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, word := range words {
		if id := m.dict.id(word); id > 0 {
			m.dict.inc(id, 1)
			continue
		}

		m.dict.add(word, 1)
	}
}

// AddWeight adds provided words to the dictionary with a custom weight
func (m *Spellchecker) AddWeight(weight uint, words ...string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, word := range words {
		if id := m.dict.id(word); id > 0 {
			m.dict.inc(id, weight)
			continue
		}

		m.dict.add(word, weight)
	}
}

var ErrUnknownWord = fmt.Errorf("unknown word")

// IsCorrect check if provided word is in the dictionary
func (s *Spellchecker) IsCorrect(word string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.dict.has(word)
}

func (s *Spellchecker) Fix(word string) (string, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.has(word) {
		return word, nil
	}

	hits := s.dict.find(word, 1)
	if len(hits) == 0 {
		return word, ErrUnknownWord
	}

	return hits[0].Value, nil
}

// Suggest find top n suggestions for the word
func (s *Spellchecker) Suggest(word string, n int) ([]string, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.has(word) {
		return []string{word}, nil
	}

	hits := s.dict.find(word, n)
	if len(hits) == 0 {
		return []string{word}, ErrUnknownWord
	}

	result := make([]string, len(hits))
	for i, h := range hits {
		result[i] = h.Value
	}

	return result, nil
}

type SuggestionResult struct {
	ExactMatch  bool
	Suggestions []Match
}

// SuggestScore find top n suggestions for the word.
// Returns spellchecker scores along with words
func (s *Spellchecker) SuggestScore(word string, n int) SuggestionResult {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.has(word) {
		return SuggestionResult{ExactMatch: true}
	}

	return SuggestionResult{
		Suggestions: s.dict.find(word, n),
	}
}
