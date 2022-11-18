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

	if maxErrors <= 0 {
		return nil
	}

	bm := d.alphabet.encode([]rune(word))
	result := make([]Match, 0, n*10)

	// exact match
	ids := d.index.get(bm)
	for _, id := range ids {
		doc, ok := d.docRaw(id)
		if !ok {
			continue
		}

		if !d.isValidWord(word, doc.Word, 0, maxErrors) {
			continue
		}

		result = append(result, Match{
			Value: doc.Word,
			Score: 0.0, // @todo calc score
		})
	}

	result = append(result, d.getFixes(word, bm, 1, maxErrors, make(map[bitmap]struct{}))...)

	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })

	if len(result) < n {
		return result
	}

	return result[0:n]
}

func (d *Dictionary) getFixes(word string, bm bitmap, errCnt int, maxErrors int, checked map[bitmap]struct{}) []Match {
	if errCnt > maxErrors {
		return nil
	}

	result := make([]Match, 0, len(d.alphabet))
	for i := 0; i < len(d.alphabet); i++ {
		bm := bm.clone()
		bm.xor(uint32(i))

		if _, ok := checked[bm]; ok {
			continue
		}
		checked[bm] = struct{}{}

		ids := d.index[bm]
		if len(ids) == 0 {
			continue
		}

		for _, id := range ids {
			doc, ok := d.docRaw(id)
			if !ok {
				continue
			}

			if !d.isValidWord(word, doc.Word, errCnt, maxErrors) {
				continue
			}

			result = append(result, Match{
				Value: doc.Word,
				Score: 0.0, // @todo calc score
			})
		}

		result = append(result, d.getFixes(word, bm, errCnt+1, maxErrors, checked)...)
	}

	return result
}

func (d *Dictionary) isValidWord(searchWord string, word string, errCnt int, maxErrors int) bool {
	searchRunes := []rune(searchWord)
	wordRunes := []rune(word)
	allowedErrs := maxErrors - errCnt - abs(len(wordRunes)-len(searchRunes))
	if allowedErrs < 0 {
		return false
	}

	if allowedErrs == 0 && searchWord != word {
		return false
	}

	// @todo levenshtein distance

	return true
}

func abs(x int) int {
	if x < 0 {
		return -1 * x
	}
	return x
}
