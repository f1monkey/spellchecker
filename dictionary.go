package spellchecker

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"math"
	"sort"
	"sync"

	"github.com/agnivade/levenshtein"
)

// maxErrros is not a "max errors" in a word. It is a max diff in bits betweeen the "search word" and a "dictionary word".
// i.e. one simple symbol replacement (problam => problem ) is a two-bit difference.
const maxErrors = 2

type Doc struct {
	Word  string
	Count int
}

type dictionary struct {
	mtx sync.RWMutex

	alphabet alphabet
	nextID   uint32
	ids      map[string]uint32
	docs     map[uint32]Doc

	index map[bitmap][]uint32
}

func newDictionary(ab Alphabet) (*dictionary, error) {
	alphabet, err := newAlphabet(ab.Letters, ab.Length)
	if err != nil {
		return nil, err
	}

	return &dictionary{
		alphabet: alphabet,
		nextID:   1,
		ids:      make(map[string]uint32),
		docs:     make(map[uint32]Doc),
		index:    make(map[bitmap][]uint32),
	}, nil
}

// id Get ID of the word. Returns 0 if not found
func (d *dictionary) id(word string) uint32 {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	return d.ids[word]
}

// has Check if word is present in the dictionary
func (d *dictionary) has(word string) bool {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	return d.ids[word] > 0
}

// doc get document by id
func (d *dictionary) doc(id uint32) (Doc, bool) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	doc, ok := d.docs[id]
	return doc, ok
}

// add Puts new word to the dictionary
func (d *dictionary) add(word string) (uint32, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	id := d.nextID
	d.ids[word] = id
	d.nextID++

	runes := []rune(word)
	d.docs[id] = Doc{Word: word, Count: 1}
	m := d.alphabet.encode(runes)
	d.index[m] = append(d.index[m], id)

	return id, nil
}

// inc Increase word occurence counter
func (d *dictionary) inc(id uint32) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	doc, ok := d.docs[id]
	if !ok {
		return
	}
	doc.Count++
	d.docs[id] = doc
}

type match struct {
	Value string
	Score float64
}

func (d *dictionary) Find(word string, n int) []match {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if maxErrors <= 0 {
		return nil
	}

	bm := d.alphabet.encode([]rune(word))
	candidates := d.getCandidates(word, bm, 1)
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

func (d *dictionary) getCandidates(word string, bmSrc bitmap, errCnt int) []сandidate {
	checked := make(map[bitmap]struct{}, d.alphabet.len()*2)

	// "exact match" OR "candidate has all the same letters as the word but in different order"
	result := make([]сandidate, 0, 50)
	if _, ok := checked[bmSrc]; !ok {
		checked[bmSrc] = struct{}{}
		ids := d.index[bmSrc]
		for _, id := range ids {
			doc, ok := d.docs[id]
			if !ok {
				continue
			}

			distance := levenshtein.ComputeDistance(word, doc.Word)
			if distance > maxErrors {
				continue
			}
			result = append(result, сandidate{
				Word:     doc.Word,
				Count:    doc.Count,
				Distance: distance,
			})
		}
	}
	// the most common mistake is a transposition of letters.
	// so if we found one here, we do early termination
	if len(result) != 0 {
		return result
	}

	// @todo perform phonetic analysis with early termination here

	// @todo try to use tree index here
	toCheck := []bitmap{bmSrc}
	for errCnt := 1; errCnt <= maxErrors; errCnt++ {
		toCheck2 := make([]bitmap, 0, d.alphabet.len())
		for _, bm := range toCheck {
			for i := 0; i < len(d.alphabet); i++ {
				bmCandidate := bm.clone()
				bmCandidate.xor(uint32(i))
				if _, ok := checked[bmCandidate]; ok {
					continue
				}
				checked[bmCandidate] = struct{}{}
				toCheck2 = append(toCheck2, bmCandidate)

				ids := d.index[bmCandidate]
				for _, id := range ids {
					doc, ok := d.docs[id]
					if !ok {
						continue
					}

					distance := levenshtein.ComputeDistance(word, doc.Word)
					if distance > maxErrors {
						continue
					}
					result = append(result, сandidate{
						Word:     doc.Word,
						Count:    doc.Count,
						Distance: distance,
					})
				}
			}
		}
		toCheck = toCheck2
	}

	return result
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
	NextID   uint32
	IDs      map[string]uint32
	Docs     map[uint32]Doc

	Counts map[uint32]int
	Index  map[bitmap][]uint32
}

func (d *dictionary) MarshalBinary() ([]byte, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	data := &dictData{
		Alphabet: d.alphabet,
		NextID:   d.nextID,
		IDs:      d.ids,
		Docs:     d.docs,
		Index:    d.index,
	}

	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (d *dictionary) UnmarshalBinary(data []byte) error {
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
	d.index = dictData.Index

	return nil
}
