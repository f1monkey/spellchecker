package spellchecker

import (
	"encoding/gob"
	"io"

	"github.com/cyradin/spellchecker-ngram/dictionary"
)

type spellcheckerData struct {
	Dict *dictionary.Dictionary
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

// Load reads spellchecker data from the provided reader and decodes it
func Load(reader io.Reader) (*Spellchecker, error) {
	data := spellcheckerData{}

	err := gob.NewDecoder(reader).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &Spellchecker{
		dict: data.Dict,
	}, nil
}
