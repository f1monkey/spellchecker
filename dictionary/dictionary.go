package dictionary

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/cyradin/ngrams"
)

const ngramSize = 3

type Doc struct {
	Value string
	Terms []string
}

var _ encoding.BinaryMarshaler = (*Dictionary)(nil)
var _ encoding.BinaryUnmarshaler = (*Dictionary)(nil)

type Index map[string]*roaring.Bitmap
type IndexByPos map[int]Index
type IndexByLen map[int]IndexByPos

type Dictionary struct {
	mtx sync.RWMutex

	nextID uint32
	ids    map[string]uint32
	docs   map[uint32]Doc

	counts  map[uint32]int
	indexes IndexByLen
}

func New() *Dictionary {
	return &Dictionary{
		nextID:  1,
		ids:     make(map[string]uint32),
		docs:    make(map[uint32]Doc),
		counts:  make(map[uint32]int),
		indexes: make(IndexByLen),
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

	return d.docRaw(id)
}

// Add Puts new word to the dictionary
func (d *Dictionary) Add(word string) (uint32, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	id := d.nextID
	d.ids[word] = id
	d.counts[id] = 1
	d.nextID++

	ngrm, err := ngrams.From(word, ngramSize)
	if err != nil {
		return 0, err
	}

	wordLen := len([]rune(word))
	d.docs[id] = Doc{Value: word, Terms: ngrm}

	for i, ng := range ngrm {
		index := d.getIndex(wordLen, i)
		m := index[ng]
		if m == nil {
			m = roaring.New()
			index[ng] = m
		}
		m.Add(id)
	}

	return id, nil
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

	Counts  map[uint32]int
	Indexes IndexByLen
}

func (d *Dictionary) MarshalBinary() ([]byte, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	data := &dictData{
		NextID:  d.nextID,
		IDs:     d.ids,
		Docs:    d.docs,
		Counts:  d.counts,
		Indexes: d.indexes,
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
	d.indexes = dictData.Indexes

	return nil
}

func (d *Dictionary) getIndex(wordLen int, pos int) Index {
	indexByPos, ok := d.indexes[wordLen]
	if !ok {
		indexByPos = make(IndexByPos)
		d.indexes[wordLen] = indexByPos
	}

	index, ok := indexByPos[pos]
	if !ok {
		index = make(Index)
		indexByPos[pos] = index
	}

	return index
}

func (d *Dictionary) docRaw(id uint32) (Doc, bool) {
	doc, ok := d.docs[id]
	return doc, ok
}
