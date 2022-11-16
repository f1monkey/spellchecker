package dictionary

import (
	"sort"
)

type Match struct {
	Value string
	Score float64
}

func (d *Dictionary) Find(word string, n int, maxErrors int) []Match {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	runes := []rune(word)
	bm := d.alphabet.encode(runes)

	return d.matchAll(bm, len(runes), maxErrors, n)

}

func (d *Dictionary) matchAll(bm bitmap, wordLen int, maxErrors int, maxCnt int) []Match {
	result := make([]Match, 0, maxCnt*10)

	// exact match
	ids := d.getIndex(wordLen).get(bm)
	for _, id := range ids {
		doc, ok := d.docRaw(id)
		if !ok {
			continue
		}

		result = append(result, Match{
			Value: doc.Word,
			Score: 0.0,
		})
	}

	for l := wordLen - maxErrors; l <= wordLen+maxErrors; l++ {
		if l <= 0 {
			continue
		}

		// @todo
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })

	if len(result) < maxCnt {
		return result
	}

	return result[0:maxCnt]
}
