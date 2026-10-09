package spellchecker

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"math"
	"sync"
	"sync/atomic"
)

type Match struct {
	Value string
	Score float64
}

type dictionary struct {
	alphabet alphabet
	zobrist  []uint64
	nextID   func() uint32

	words  map[uint32][]rune
	ids    map[string]uint32
	counts map[uint32]uint

	index map[uint64][]uint32

	pool *keyPool
}

func newDictionary(ab ...Alphabet) (*dictionary, error) {
	alphabet, err := newAlphabet(ab...)
	if err != nil {
		return nil, err
	}

	return &dictionary{
		alphabet: alphabet,
		zobrist:  alphabet.zobrist(),
		nextID:   idSeq(0),
		ids:      make(map[string]uint32),
		words:    make(map[uint32][]rune),
		counts:   make(map[uint32]uint),
		index:    make(map[uint64][]uint32),
		pool:     newKeyPool(),
	}, nil
}

// ID get ID of the word. Returns 0 if not found
func (d *dictionary) ID(word string) uint32 {
	return d.ids[word]
}

// Has check if the word is present in the dictionary
func (d *dictionary) Has(word string) bool {
	return d.ids[word] > 0
}

// Add puts the word to the dictionary
func (d *dictionary) Add(word string, n uint) uint32 {
	id := d.nextID()
	d.ids[word] = id

	wordRunes := []rune(word)

	d.counts[id] = n
	d.words[id] = wordRunes
	d.addToIndex(id, wordRunes)

	return id
}

// inc increase word occurrence counter
func (d *dictionary) Inc(id uint32, n uint) {
	_, ok := d.counts[id]
	if !ok {
		return
	}

	d.counts[id] += n
}

func (d *dictionary) Find(word string, n int, maxErrors int, fn ScoringFunc) []Match {
	if maxErrors <= 0 {
		return nil
	}

	result := newPriorityQueue(n)

	wordRunes := []rune(word)
	srcKey := d.alphabet.key(wordRunes, d.zobrist)

	// check for transposition or exact match and do early termination if found
	// (the most common mistake is a transposition of letters)
	d.fillWithCandidates(result, wordRunes, srcKey, maxErrors, fn)

	if result.Len() != 0 {
		return result.DrainSorted()
	}

	keys := d.pool.Get()
	defer d.pool.Set(keys)

	d.computeCandidateKeys(keys, srcKey, maxErrors)

	for key := range keys {
		d.fillWithCandidates(result, wordRunes, key, maxErrors, fn)
	}

	return result.DrainSorted()
}

func (d *dictionary) addToIndex(id uint32, word []rune) {
	key := d.alphabet.key(word, d.zobrist)
	d.index[key] = append(d.index[key], id)
}

// computeCandidateKeys collects index keys of letter sets that differ from src
// by at most maxFlips symbols. Flipping a symbol is a XOR with its Zobrist value.
func (d *dictionary) computeCandidateKeys(keys map[uint64]struct{}, src uint64, maxFlips int) {
	var dfs func(key uint64, level, start int)

	dfs = func(key uint64, level, start int) {
		if len(d.index[key]) > 0 {
			keys[key] = struct{}{}
		}

		if level == maxFlips {
			return
		}

		for i := start; i < len(d.zobrist); i++ {
			dfs(key^d.zobrist[i], level+1, i+1)
		}
	}

	dfs(src, 0, 0)
}

func (d *dictionary) fillWithCandidates(result *priorityQueue, wordRunes []rune, key uint64, maxErrors int, filter ScoringFunc) {
	ids := d.index[key]
	for _, id := range ids {
		docWord, ok := d.words[id]
		if !ok {
			continue
		}

		if math.Abs(float64(len(docWord)-len(wordRunes))) > float64(maxErrors) {
			continue
		}

		score, ok := filter(wordRunes, docWord, d.counts[id], maxErrors)
		if !ok {
			continue
		}

		result.Offer(score, docWord)
	}
}

var _ encoding.BinaryMarshaler = (*dictionary)(nil)
var _ encoding.BinaryUnmarshaler = (*dictionary)(nil)

// dictData is the serialized form of the dictionary.
// The index is not stored: it is rebuilt from Words on load.
type dictData struct {
	Alphabet alphabet
	IDs      map[string]uint32
	Words    map[uint32][]rune
	Counts   map[uint32]uint
}

func (d *dictionary) MarshalBinary() ([]byte, error) {
	data := &dictData{
		Alphabet: d.alphabet,
		IDs:      d.ids,
		Words:    d.words,
		Counts:   d.counts,
	}

	buf := &bytes.Buffer{}

	if err := gob.NewEncoder(buf).Encode(data); err != nil {
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
	d.zobrist = d.alphabet.zobrist()
	d.ids = dictData.IDs
	d.counts = dictData.Counts
	d.words = dictData.Words
	d.pool = newKeyPool()

	var max uint32
	for _, id := range d.ids {
		if id > max {
			max = id
		}
	}

	d.nextID = idSeq(max)

	// rebuild in id order, so buckets keep the insertion order
	d.index = make(map[uint64][]uint32)
	for id := uint32(1); id <= max; id++ {
		if word, ok := d.words[id]; ok {
			d.addToIndex(id, word)
		}
	}

	return nil
}

func idSeq(start uint32) func() uint32 {
	return func() uint32 {
		return atomic.AddUint32(&start, 1)
	}
}

type keyPool struct {
	inner *sync.Pool
}

func newKeyPool() *keyPool {
	return &keyPool{
		inner: &sync.Pool{
			New: func() any {
				return make(map[uint64]struct{}, 256)
			},
		},
	}
}

func (p *keyPool) Get() map[uint64]struct{} {
	return p.inner.Get().(map[uint64]struct{}) //nolint:forcetypeassert
}

func (p *keyPool) Set(value map[uint64]struct{}) {
	for k := range value {
		delete(value, k)
	}

	p.inner.Put(value)
}
