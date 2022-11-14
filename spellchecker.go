package spellchecker

import (
	"bufio"
	"io"
	"sync"

	dictionary "github.com/cyradin/spellchecker/dictionary"
)

// OptionFunc option setter
type OptionFunc func(m *Spellchecker) error

type Spellchecker struct {
	mtx sync.RWMutex

	dict     *dictionary.Dictionary
	splitter bufio.SplitFunc
}

func New(opts ...OptionFunc) (*Spellchecker, error) {
	dict, err := dictionary.New(dictionary.DefaultAlphabet)
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
		if id := m.dict.ID(word); id > 0 {
			m.dict.Inc(id)
			continue
		}

		m.dict.Add(word)
	}
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
