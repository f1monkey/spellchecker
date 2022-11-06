package spellchecker

import (
	"bufio"
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
	result := &Spellchecker{
		dict: newDictionary(),
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

		m.dict.add(word, makeTerms(word))
	}
}

// IsCorrect check if provided word is in the dictionary
func (s *Spellchecker) IsCorrect(word string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	id := s.dict.id(word)

	return id > 0
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
