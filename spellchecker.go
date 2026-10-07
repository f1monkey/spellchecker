package spellchecker

import (
	"sync"
)

type Spellchecker struct {
	mtx sync.RWMutex

	tokenizer Tokenizer
	dict      *dictionary
}

// New creates a spellchecker with the given tokenizer and alphabets.
// Alphabets are the allowed characters for indexing and lookup; other
// characters are ignored. Pass one or more constants such as EN, RU, Numbers,
// or any custom string.
func New(tokenizer Tokenizer, alphabets ...Alphabet) (*Spellchecker, error) {
	dict, err := newDictionary(alphabets...)
	if err != nil {
		return nil, err
	}

	result := &Spellchecker{
		tokenizer: tokenizer,
		dict:      dict,
	}

	return result, nil
}

// IsCorrect check if provided word is in the dictionary
func (s *Spellchecker) IsCorrect(word string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.dict.has(word)
}

// Add adds each argument as a whole dictionary word. Strings are not split;
// use AddPhrases to tokenize text.
func (s *Spellchecker) Add(words ...string) {
	s.AddWeight(1, words...)
}

// AddWeight is like Add, but each word is stored with the given frequency
// weight. If the word is already in the dictionary, the weight is added to
// its count. Higher weight ranks the word higher in suggestions.
func (s *Spellchecker) AddWeight(weight uint, words ...string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, word := range words {
		if word == "" {
			continue
		}

		if id := s.dict.id(word); id > 0 {
			s.dict.inc(id, weight)

			continue
		}

		s.dict.add(word, weight)
	}
}

// AddPhrases tokenizes each phrase with the spellchecker's tokenizer and adds
// the resulting words with weight 1.
func (s *Spellchecker) AddPhrases(phrases ...string) {
	s.AddPhraseWeight(1, phrases...)
}

// AddPhraseWeight is like AddPhrases, but each token is added with the given
// weight.
func (s *Spellchecker) AddPhraseWeight(weight uint, phrases ...string) {
	for _, phrase := range phrases {
		for _, word := range s.tokenizer.Tokenize(phrase) {
			s.AddWeight(weight, word)
		}
	}
}
