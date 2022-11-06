package spellchecker

import "fmt"

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

	if s.dict.Has(word) {
		return []string{word}, nil
	}

	hits := s.dict.Find(word, 1)
	if len(hits) == 0 {
		return []string{word}, fmt.Errorf("%w: %s", ErrUnknownWord, word)
	}

	result := make([]string, len(hits))
	for i, h := range hits {
		result[i] = h.Value
	}

	return result, nil
}
