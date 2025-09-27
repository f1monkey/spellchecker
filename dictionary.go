package spellchecker

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"sort"
	"sync/atomic"

	"github.com/f1monkey/bitmap"
)

type dictionary struct {
	maxErrors int
	alphabet  alphabet
	nextID    func() uint32

	words  map[uint32]string
	ids    map[string]uint32
	counts map[uint32]uint

	index map[uint64][]uint32

	filterFunc FilterFunc
}

func newDictionary(ab string, filterFunc FilterFunc, maxErrors int) (*dictionary, error) {
	alphabet, err := newAlphabet(ab)
	if err != nil {
		return nil, err
	}

	return &dictionary{
		maxErrors:  maxErrors,
		alphabet:   alphabet,
		nextID:     idSeq(0),
		ids:        make(map[string]uint32),
		words:      make(map[uint32]string),
		counts:     make(map[uint32]uint),
		index:      make(map[uint64][]uint32),
		filterFunc: filterFunc,
	}, nil
}

// id get ID of the word. Returns 0 if not found
func (d *dictionary) id(word string) uint32 {
	return d.ids[word]
}

// has check if the word is present in the dictionary
func (d *dictionary) has(word string) bool {
	return d.ids[word] > 0
}

// add puts the word to the dictionary
func (d *dictionary) add(word string, n uint) (uint32, error) {
	id := d.nextID()
	d.ids[word] = id

	d.counts[id] = n
	d.words[id] = word
	key := sum(d.alphabet.encode([]rune(word)))
	d.index[key] = append(d.index[key], id)

	return id, nil
}

// inc increase word occurence counter
func (d *dictionary) inc(id uint32, n uint) {
	_, ok := d.counts[id]
	if !ok {
		return
	}
	d.counts[id] += n
}

type Match struct {
	Value string
	Score float64
}

func (d *dictionary) find(word string, n int) []Match {
	if d.maxErrors <= 0 {
		return nil
	}

	candidates := d.getCandidates(word, n)
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Score > candidates[j].Score })

	return candidates
}

func (d *dictionary) getCandidates(word string, max int) []Match {
	result := newPriorityQueue(max)

	wordRunes := []rune(word)
	bmSrc := d.alphabet.encode([]rune(wordRunes))

	// check for transposition or exact match and do early termination if found
	// (the most common mistake is a transposition of letters)
	d.fillWithCandidates(result, wordRunes, sum(bmSrc))
	if result.Len() != 0 {
		return result.items
	}

	for bm := range d.computeCandidateBitmaps(bmSrc, d.maxErrors) {
		d.fillWithCandidates(result, wordRunes, bm)
	}

	return result.items
}

func (d *dictionary) computeCandidateBitmaps(bmSrc bitmap.Bitmap32, maxFlips int) map[uint64]struct{} {
	bitmaps := make(map[uint64]struct{}, d.alphabet.len()*maxFlips*2)

	var dfs func(bm bitmap.Bitmap32, level, start int)
	dfs = func(bm bitmap.Bitmap32, level, start int) {
		key := sum(bm)
		if len(d.index[key]) > 0 {
			bitmaps[key] = struct{}{}
		}

		if level == maxFlips {
			return
		}

		for i := start; i < d.alphabet.len(); i++ {
			bm.Xor(uint32(i)) // change one bit
			dfs(bm, level+1, i+1)
			bm.Xor(uint32(i)) // return back the changed bit
		}
	}

	dfs(bmSrc.Clone(), 0, 0)

	return bitmaps
}

func (d *dictionary) fillWithCandidates(result *priorityQueue, wordRunes []rune, bm uint64) {
	ids := d.index[bm]
	for _, id := range ids {
		docWord, ok := d.words[id]
		if !ok {
			continue
		}

		score, ok := d.filterFunc(wordRunes, []rune(docWord), d.counts[id])
		if !ok {
			continue
		}

		result.Push(Match{
			Value: docWord,
			Score: score,
		})
	}
}

var _ encoding.BinaryMarshaler = (*dictionary)(nil)
var _ encoding.BinaryUnmarshaler = (*dictionary)(nil)

type dictData struct {
	Alphabet alphabet
	IDs      map[string]uint32
	Words    map[uint32]string
	Counts   map[uint32]uint

	Index map[uint64][]uint32

	MaxErrors int
}

func (d *dictionary) MarshalBinary() ([]byte, error) {
	data := &dictData{
		Alphabet:  d.alphabet,
		IDs:       d.ids,
		Words:     d.words,
		Counts:    d.counts,
		Index:     d.index,
		MaxErrors: d.maxErrors,
	}

	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (d *dictionary) UnmarshalBinary(data []byte) error {
	dictData := &dictData{}
	err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(dictData)
	if err != nil {
		return err
	}

	d.alphabet = dictData.Alphabet
	d.ids = dictData.IDs
	d.counts = dictData.Counts
	d.words = dictData.Words
	d.index = dictData.Index
	d.maxErrors = dictData.MaxErrors
	d.filterFunc = defaultFilterFunc(dictData.MaxErrors)

	var max uint32
	for _, id := range d.ids {
		if id > max {
			max = id
		}
	}
	d.nextID = idSeq(max)

	return nil
}

func idSeq(start uint32) func() uint32 {
	return func() uint32 {
		return atomic.AddUint32(&start, 1)
	}
}

func sum(b bitmap.Bitmap32) uint64 {
	var result uint64
	var mult uint64 = 1
	for i := range b {
		result += uint64(b[i]) * mult
		mult *= 10
	}

	return result
}
