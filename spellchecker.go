package spellchecker

import (
	"bufio"
	"fmt"
	"io"
	"sync"
)

// OptionFunc option setter
type OptionFunc func(m *Spellchecker) error

type Spellchecker struct {
	mtx sync.RWMutex

	dict     *dictionary
	splitter bufio.SplitFunc
}

func New(opts ...OptionFunc) (*Spellchecker, error) {
	dict, err := newDictionary(DefaultAlphabet)
	if err != nil {
		return nil, err
	}

	result := &Spellchecker{
		dict: dict,
	}
	for _, o := range opts {
		if err := o(result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// AddFrom reads input, splits it with spellchecker splitter func and adds words to dictionary
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

// Add adds provided words to dictionary
func (m *Spellchecker) Add(words ...string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, word := range words {
		if id := m.dict.id(word); id > 0 {
			m.dict.inc(id)
			continue
		}

		m.dict.add(word)
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

	hits := s.dict.Find(word, 1)
	if len(hits) == 0 {
		return word, fmt.Errorf("%w: %s", ErrUnknownWord, word)
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

	hits := s.dict.Find(word, n)
	if len(hits) == 0 {
		return []string{word}, fmt.Errorf("%w: %s", ErrUnknownWord, word)
	}

	result := make([]string, len(hits))
	for i, h := range hits {
		result[i] = h.Value
	}

	return result, nil
}

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
