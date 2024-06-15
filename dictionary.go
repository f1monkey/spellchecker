package spellchecker

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"math"
	"sort"
	"sync/atomic"

	"github.com/agnivade/levenshtein"
	"github.com/f1monkey/bitmap"
)

type dictionary struct {
	maxErrors int
	alphabet  alphabet
	nextID    func() uint32

	words  map[uint32]string
	ids    map[string]uint32
	counts map[uint32]int

	index map[uint64][]uint32
}

func newDictionary(ab string, maxErrors int) (*dictionary, error) {
	alphabet, err := newAlphabet(ab)
	if err != nil {
		return nil, err
	}

	return &dictionary{
		maxErrors: maxErrors,
		alphabet:  alphabet,
		nextID:    idSeq(0),
		ids:       make(map[string]uint32),
		words:     make(map[uint32]string),
		counts:    make(map[uint32]int),
		index:     make(map[uint64][]uint32),
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
func (d *dictionary) add(word string) (uint32, error) {
	id := d.nextID()
	d.ids[word] = id

	runes := []rune(word)
	d.counts[id] = 1
	d.words[id] = word
	key := sum(d.alphabet.encode(runes))
	d.index[key] = append(d.index[key], id)

	return id, nil
}

// inc increase word occurence counter
func (d *dictionary) inc(id uint32) {
	_, ok := d.counts[id]
	if !ok {
		return
	}
	d.counts[id]++
}

type match struct {
	Value string
	Score float64
}

func (d *dictionary) find(word string, n int) []match {
	if d.maxErrors <= 0 {
		return nil
	}

	bm := d.alphabet.encode([]rune(word))
	candidates := d.getCandidates(word, bm)
	result := calcScores([]rune(word), candidates)

	if len(result) < n {
		return result
	}

	return result[0:n]
}

type сandidate struct {
	Word     string
	Distance int
	Count    int
}

func (d *dictionary) getCandidates(word string, bmSrc bitmap.Bitmap32) []сandidate {
	checked := make(map[uint64]struct{}, d.alphabet.len()*2)

	result := make([]сandidate, 0, 50)

	// "exact match" OR "candidate has all the same letters as the word but in different order"
	key := sum(bmSrc)
	checked[key] = struct{}{}
	ids := d.index[key]
	for _, id := range ids {
		docWord, ok := d.words[id]
		if !ok {
			continue
		}

		distance := levenshtein.ComputeDistance(word, docWord)
		if distance > d.maxErrors {
			continue
		}
		result = append(result, сandidate{
			Word:     docWord,
			Count:    d.counts[id],
			Distance: distance,
		})
	}
	// the most common mistake is a transposition of letters.
	// so if we found one here, we do early termination
	if len(result) != 0 {
		return result
	}

	// @todo perform phonetic analysis with early termination here
	for bm := range d.computeCandidateBitmaps(bmSrc) {
		ids := d.index[bm]
		for _, id := range ids {
			docWord, ok := d.words[id]
			if !ok {
				continue
			}

			distance := levenshtein.ComputeDistance(word, docWord)
			if distance > d.maxErrors {
				continue
			}
			result = append(result, сandidate{
				Word:     docWord,
				Count:    d.counts[id],
				Distance: distance,
			})
		}
	}

	return result
}

func (d *dictionary) computeCandidateBitmaps(bmSrc bitmap.Bitmap32) map[uint64]struct{} {
	bitmaps := make(map[uint64]struct{}, d.alphabet.len()*5)

	// swap one bit
	for i := 0; i < d.alphabet.len(); i++ {
		bit := uint32(i)
		bmCandidate := bmSrc.Clone()
		bmCandidate.Xor(bit)

		// swap one more bit to be able to fix:
		// - two deletions ("rang" => "orange")
		// - replacements ("problam" => "problem")
		for j := 0; j < d.alphabet.len(); j++ {
			bit := uint32(j)
			bmCandidate := bmCandidate.Clone()
			bmCandidate.Xor(bit)
			key := sum(bmCandidate)
			if len(d.index[key]) == 0 {
				continue
			}
			bitmaps[key] = struct{}{}
		}

		key := sum(bmCandidate)
		if len(d.index[key]) == 0 {
			continue
		}
		bitmaps[key] = struct{}{}
	}

	return bitmaps
}

func calcScores(src []rune, candidates []сandidate) []match {
	result := make([]match, len(candidates))
	for i, c := range candidates {
		result[i] = match{
			Value: c.Word,
			Score: calcScore(src, []rune(c.Word), c.Distance, c.Count),
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })

	return result
}

func calcScore(src []rune, candidate []rune, distance int, cnt int) float64 {
	mult := math.Log1p(float64(cnt))
	// if first letters are the same, increase score
	if src[0] == candidate[0] {
		mult *= 1.5
		// if second letters are the same too, increase score even more
		if len(src) > 1 && len(candidate) > 1 && src[1] == candidate[1] {
			mult *= 1.5
		}
	}

	return 1 / (1 + float64(distance*distance)) * mult
}

var _ encoding.BinaryMarshaler = (*dictionary)(nil)
var _ encoding.BinaryUnmarshaler = (*dictionary)(nil)

type dictData struct {
	Alphabet alphabet
	IDs      map[string]uint32
	Words    map[uint32]string
	Counts   map[uint32]int

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
