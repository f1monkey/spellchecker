package spellchecker

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
)

type AddOptionFunc func(opts *addOptions)

// AddWithWeight sets weight for added words.
// The weight increases the likelihood that the word will be chosen as a correction.
func AddWithWeight(weight uint) AddOptionFunc {
	return func(opts *addOptions) {
		opts.weight = weight
	}
}

// AddWithSplitter sets a splitter func for AddFrom() reader
func AddWithSplitter(splitter bufio.SplitFunc) AddOptionFunc {
	return func(opts *addOptions) {
		opts.splitter = splitter
	}
}

// AddFrom reads input, splits it with spellchecker splitter func and adds words to the dictionary
func (m *Spellchecker) AddFrom(input io.Reader, opts ...AddOptionFunc) error {
	addOpts := defaultAddOptions
	for _, o := range opts {
		o(&addOpts)
	}

	words := make([]string, 1000)
	i := 0
	for item := range readInput(input, addOpts.splitter) {
		if item.err != nil {
			return item.err
		}

		if i == len(words) {
			m.addMany(words, addOpts.weight)
			i = 0
		}
		words[i] = item.word
		i++
	}

	if i > 0 {
		m.addMany(words, addOpts.weight)
	}

	return nil
}

// AddMany adds provided words to the dictionary
func (m *Spellchecker) AddMany(words []string, opts ...AddOptionFunc) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	addOpts := defaultAddOptions
	for _, o := range opts {
		o(&addOpts)
	}

	m.addMany(words, addOpts.weight)
}

// Add adds provided word to the dictionary
func (m *Spellchecker) Add(word string, opts ...AddOptionFunc) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	addOpts := defaultAddOptions
	for _, o := range opts {
		o(&addOpts)
	}

	m.add(word, addOpts.weight)
}

func (m *Spellchecker) addMany(words []string, weight uint) {
	for _, word := range words {
		m.add(word, weight)
	}
}

func (m *Spellchecker) add(word string, weight uint) {
	if id := m.dict.id(word); id > 0 {
		m.dict.inc(id, weight)
		return
	}

	m.dict.add(word, weight)
}

type addOptions struct {
	weight   uint
	splitter bufio.SplitFunc
}

var defaultAddOptions = addOptions{
	weight:   1,
	splitter: defaultSplitter,
}

var wordSymbols = regexp.MustCompile(`[-\pL]+`)

func defaultSplitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanWords(data, atEOF)
	if err != nil {
		return
	}
	token = bytes.ToLower(token)

	return advance, wordSymbols.Find(token), nil
}
