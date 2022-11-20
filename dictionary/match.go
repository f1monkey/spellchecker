package dictionary

import (
	"math"
	"sort"

	"github.com/agnivade/levenshtein"
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

		if score := d.calcScore(word, doc.Word, 0, maxErrors, doc.Count); score != 0 {
			result = append(result, Match{
				Value: doc.Word,
				Score: score,
			})
		}
	}

	checked := make(map[bitmap]struct{})
	result = append(result, d.getFixes(word, bm, 1, maxErrors, checked)...)

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
		if len(ids) != 0 {
			for _, id := range ids {
				doc, ok := d.docRaw(id)
				if !ok {
					continue
				}

				if score := d.calcScore(word, doc.Word, 0, maxErrors, doc.Count); score != 0 {
					result = append(result, Match{
						Value: doc.Word,
						Score: score,
					})
				}
			}

		}

		result = append(result, d.getFixes(word, bm, errCnt+1, maxErrors, checked)...)
	}

	return result
}

func (d *Dictionary) calcScore(searchWord string, word string, errCnt int, maxErrors int, count int) float64 {
	searchRunes := []rune(searchWord)
	wordRunes := []rune(word)
	allowedErrs := maxErrors - errCnt - abs(len(wordRunes)-len(searchRunes))
	if allowedErrs < 0 {
		return 0.0
	}

	if allowedErrs == 0 && searchWord != word {
		return 0.0
	}

	mult := 1 / (1 + float64(errCnt*errCnt)) * math.Log1p(float64(count))

	// if first letters are the same, increase score
	if searchRunes[0] == wordRunes[0] {
		mult *= 1.5
		// if second letters are the same too, increase score even more
		if len(searchRunes) > 1 && len(wordRunes) > 1 && searchRunes[1] == wordRunes[1] {
			mult *= 1.5
		}
	}

	distance := levenshtein.ComputeDistance(searchWord, word)

	return 1 / (1 + float64(distance*distance)) * mult
}

func abs(x int) int {
	if x < 0 {
		return -1 * x
	}
	return x
}
