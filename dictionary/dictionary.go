package dictionary

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"sync"
)

const ngramSize = 3

type Doc struct {
	Word  string
	Count int
}

var _ encoding.BinaryMarshaler = (*Dictionary)(nil)
var _ encoding.BinaryUnmarshaler = (*Dictionary)(nil)

type Dictionary struct {
	mtx sync.RWMutex

	alphabet alphabet
	nextID   uint32
	ids      map[string]uint32
	docs     map[uint32]Doc

	indexes IndexByLen
}

func New(ab string) (*Dictionary, error) {
	alphabet, err := newAlphabet(ab)
	if err != nil {
		return nil, err
	}

	return &Dictionary{
		alphabet: alphabet,
		nextID:   1,
		ids:      make(map[string]uint32),
		docs:     make(map[uint32]Doc),
		indexes:  make(IndexByLen),
	}, nil
}

// ID Get ID of the word. Returns 0 if not found
func (d *Dictionary) ID(word string) uint32 {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	return d.ids[word]
}

// Has Check if word is present in the dictionary
func (d *Dictionary) Has(word string) bool {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	return d.ids[word] > 0
}

// Doc get document by id
func (d *Dictionary) Doc(id uint32) (Doc, bool) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	return d.docRaw(id)
}

// Add Puts new word to the dictionary
func (d *Dictionary) Add(word string) (uint32, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	id := d.nextID
	d.ids[word] = id
	d.nextID++

	runes := []rune(word)
	wordLen := len(runes)
	d.docs[id] = Doc{Word: word, Count: 1}
	index := d.getIndex(wordLen)
	m := d.alphabet.encode(runes)
	index[m] = append(index[m], id)

	return id, nil
}

// Inc Increase word occurence counter
func (d *Dictionary) Inc(id uint32) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	doc, ok := d.docRaw(id)
	if !ok {
		return
	}
	doc.Count++
	d.docs[id] = doc
}

type dictData struct {
	Alphabet alphabet
	NextID   uint32
	IDs      map[string]uint32
	Docs     map[uint32]Doc

	Counts  map[uint32]int
	Indexes IndexByLen
}

func (d *Dictionary) MarshalBinary() ([]byte, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	data := &dictData{
		Alphabet: d.alphabet,
		NextID:   d.nextID,
		IDs:      d.ids,
		Docs:     d.docs,
		Indexes:  d.indexes,
	}

	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (d *Dictionary) UnmarshalBinary(data []byte) error {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	dictData := &dictData{}
	err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(dictData)
	if err != nil {
		return err
	}

	d.alphabet = dictData.Alphabet
	d.nextID = dictData.NextID
	d.ids = dictData.IDs
	d.docs = dictData.Docs
	d.indexes = dictData.Indexes

	return nil
}

func (d *Dictionary) getIndex(wordLen int) Index {
	index, ok := d.indexes[wordLen]
	if !ok {
		index = make(Index)
		d.indexes[wordLen] = index
	}

	return index
}

func (d *Dictionary) docRaw(id uint32) (Doc, bool) {
	doc, ok := d.docs[id]
	return doc, ok
}
