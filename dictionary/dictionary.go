package dictionary

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/cyradin/spellchecker/ngram"
)

const MaxNGRam = 5

type Doc struct {
	Value string
	Terms []ngram.NGram
}

var _ encoding.BinaryMarshaler = (*Dictionary)(nil)
var _ encoding.BinaryUnmarshaler = (*Dictionary)(nil)

type Dictionary struct {
	mtx sync.RWMutex

	nextID uint32
	ids    map[string]uint32
	docs   map[uint32]Doc

	counts map[uint32]int
	index  map[string]*roaring.Bitmap
}

func New() *Dictionary {
	return &Dictionary{
		nextID: 1,
		ids:    make(map[string]uint32),
		docs:   make(map[uint32]Doc),
		counts: make(map[uint32]int),
		index:  make(map[string]*roaring.Bitmap),
	}
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

	doc, ok := d.docs[id]
	return doc, ok
}

// Add Puts new word to the dictionary
func (d *Dictionary) Add(word string) uint32 {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	id := d.nextID
	d.ids[word] = id
	d.counts[id] = 1
	d.nextID++

	tt := ngram.MakeAll(word, MaxNGRam)
	d.docs[id] = Doc{Value: word, Terms: tt}

	for _, t := range tt {
		m := d.index[t.Value]
		if m == nil {
			m = roaring.New()
			d.index[t.Value] = m
		}
		m.Add(id)
	}

	return id
}

// Inc Increase word occurence counter
func (d *Dictionary) Inc(id uint32) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	d.counts[id]++
}

type dictData struct {
	NextID uint32
	IDs    map[string]uint32
	Docs   map[uint32]Doc

	Counts map[uint32]int
	Index  map[string]*roaring.Bitmap
}

func (d *Dictionary) MarshalBinary() ([]byte, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	data := &dictData{
		NextID: d.nextID,
		IDs:    d.ids,
		Docs:   d.docs,
		Counts: d.counts,
		Index:  d.index,
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

	d.nextID = dictData.NextID
	d.ids = dictData.IDs
	d.docs = dictData.Docs
	d.counts = dictData.Counts
	d.index = dictData.Index

	return nil
}
