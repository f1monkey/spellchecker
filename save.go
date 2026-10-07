package spellchecker

import (
	"encoding/gob"
	"io"
)

type spellcheckerData struct {
	Dict *dictionary
}

// Save encodes spellchecker data and writes it to the provided writer
func (m *Spellchecker) Save(w io.Writer) error {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	data := spellcheckerData{
		Dict: m.dict,
	}

	return gob.NewEncoder(w).Encode(data)
}

// Load reads spellchecker data from the provided reader and decodes it.
// tokenizer is used for AddPhrases after loading.
func Load(reader io.Reader, tokenizer Tokenizer) (*Spellchecker, error) {
	data := spellcheckerData{}

	err := gob.NewDecoder(reader).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &Spellchecker{
		tokenizer: tokenizer,
		dict:      data.Dict,
	}, nil
}
