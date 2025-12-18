package spellchecker

import (
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

// IsCorrect check if provided word is in the dictionary
func (s *Spellchecker) IsCorrect(word string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.dict.has(word)
}
