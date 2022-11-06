package dictionary

import (
	"github.com/RoaringBitmap/roaring"
)

type Doc struct {
	Value string
}

type Dictionary struct {
	NextID uint32
	IDs    map[string]uint32
	Docs   map[uint32]Doc

	Counts map[uint32]int
	Index  map[string]*roaring.Bitmap
}

func New() *Dictionary {
	return &Dictionary{
		NextID: 1,
		IDs:    make(map[string]uint32),
		Docs:   make(map[uint32]Doc),
		Counts: make(map[uint32]int),
		Index:  make(map[string]*roaring.Bitmap),
	}
}

// ID Get ID of the word. Returns 0 if not found
func (d *Dictionary) ID(word string) uint32 {
	return d.IDs[word]
}

// Has Check if word is present in the dictionary
func (d *Dictionary) Has(word string) bool {
	return d.IDs[word] > 0
}

// Add Puts new word to the dictionary
func (d *Dictionary) Add(word string) uint32 {
	id := d.NextID
	d.IDs[word] = id
	d.Counts[id] = 1
	d.NextID++

	for _, t := range makeTerms(word) {
		m := d.Index[t.Value]
		if m == nil {
			m = roaring.New()
			d.Index[t.Value] = m
		}
		m.Add(id)
	}

	return id
}

// Inc Increase word occurence counter
func (d *Dictionary) Inc(id uint32) {
	d.Counts[id]++
}
