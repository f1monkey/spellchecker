package spellchecker

import (
	"fmt"

	"github.com/cyradin/spellchecker/dictionary"
	"github.com/cyradin/spellchecker/ngram"
)

// MaxEditsAuto word length from 0 to 2: 0 edits; from 3 to 5: 1 edit; > 5: 2 edits
const MaxEditsAuto = -1

var ErrUnknownWord = fmt.Errorf("unknown word")

// IsCorrect check if provided word is in the dictionary
func (s *Spellchecker) IsCorrect(word string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.dict.Has(word)
}

func (s *Spellchecker) Fix(word string) (string, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.Has(word) {
		return word, nil
	}

	hits := s.find(word, 1)
	if len(hits) == 0 {
		return word, fmt.Errorf("%w: %s", ErrUnknownWord, word)
	}

	return hits[0].Value, nil
}

// Suggest find top n suggestions for the word
func (s *Spellchecker) Suggest(word string, n int) ([]string, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.dict.Has(word) {
		return []string{word}, nil
	}

	hits := s.find(word, 1)
	if len(hits) == 0 {
		return []string{word}, fmt.Errorf("%w: %s", ErrUnknownWord, word)
	}

	result := make([]string, len(hits))
	for i, h := range hits {
		result[i] = h.Value
	}

	return result, nil
}

type Hit struct {
	Value string
	Score float64
}

// find returns top N hits by word
func (s *Spellchecker) find(word string, n int) []Hit {
	tt := ngram.Make(word, dictionary.MaxNGRam)
	matches := s.dict.Match(tt)
	if matches.IsEmpty() {
		return nil
	}

	// @todo calculate scores and return top hits
	result := make([]Hit, 0, 20)

	matches.Iterate(func(x uint32) bool {
		doc, ok := s.dict.Doc(x)
		if !ok {
			return true
		}

		result = append(result, Hit{
			Value: doc.Value,
		})

		return true
	})

	return result
}
