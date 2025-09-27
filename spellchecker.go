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

	dict, err := newDictionary(alphabet, result.maxErrors)
	if err != nil {
		return nil, err
	}

	result.dict = dict

	return result, nil
}

// AddFrom reads input, splits it with spellchecker splitter func and adds words to the dictionary
func (m *Spellchecker) AddFrom(weight uint, input io.Reader) error {
	words := make([]string, 1000)
	i := 0
	for item := range readInput(input, m.splitter) {
		if item.err != nil {
			return item.err
		}

		if i == len(words) {
			m.Add(weight, words...)
			i = 0
		}
		words[i] = item.word
		i++
	}

	if i > 0 {
		m.Add(weight, words[:i]...)
	}

	return nil
}

// Add adds provided words to the dictionary with a custom weight
func (m *Spellchecker) Add(weight uint, words ...string) {
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

	hits := s.dict.find(word, 1, s.filterFunc)
	if len(hits) == 0 {
		return word, ErrUnknownWord
	}

	return hits[0].Value, nil
}

type SuggestionResult struct {
	ExactMatch  bool // if true, the word is correct
	Suggestions []Match
}

// Suggest find top n suggestions for the word.
// Returns spellchecker scores along with words
func (s *Spellchecker) Suggest(word string, n int) SuggestionResult {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.has(word) {
		return SuggestionResult{ExactMatch: true}
	}

	return SuggestionResult{
		Suggestions: s.dict.find(word, n, s.filterFunc),
	}
}
