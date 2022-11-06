package spellchecker

import (
	"github.com/RoaringBitmap/roaring"
)

type dictionary struct {
	NextID uint32
	IDs    map[string]uint32
	Counts map[uint32]int
	Index  map[string]*roaring.Bitmap
}

func newDictionary() *dictionary {
	return &dictionary{
		NextID: 1,
		IDs:    make(map[string]uint32),
		Counts: make(map[uint32]int),
		Index:  make(map[string]*roaring.Bitmap),
	}
}

func (d *dictionary) id(word string) uint32 {
	return d.IDs[word]
}

func (d *dictionary) add(word string, terms []Term) uint32 {
	id := d.NextID
	d.IDs[word] = id
	d.Counts[id] = 1
	d.NextID++

	for _, t := range terms {
		m := d.Index[t.Value]
		if m == nil {
			m = roaring.New()
			d.Index[t.Value] = m
		}
		m.Add(id)
	}

	return id
}

func (d *dictionary) inc(id uint32) {
	d.Counts[id]++
}
