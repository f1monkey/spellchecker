package dictionary

import (
	"math"
	"sort"

	"github.com/RoaringBitmap/roaring"
	"github.com/cyradin/ngrams"
)

type Match struct {
	Value string
	Score float64
}

func (d *Dictionary) Find(word string, n int, maxErrors int) []Match {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	ngrms, _ := ngrams.From(word, ngramSize)
	if len(ngrms) == 0 {
		return nil
	}
	wordLen := len([]rune(word))

	return d.matchAll(ngrms, wordLen, maxErrors, n)
}

type matchOpts struct {
	realLength  int
	matchLength int
	ngrm        []string
	offset      int
	maxErrors   int
}

func (d *Dictionary) matchAll(ngrms []string, wordLen int, maxErrors int, maxCnt int) []Match {
	result := make([]Match, 0, maxCnt*10)

	// match with different lengths
	for l := wordLen - maxErrors; l <= wordLen+maxErrors; l++ {
		lengthErrs := abs(wordLen - maxErrors)

		ngrmMap := make(map[string]struct{})
		for _, ng := range ngrms {
			ngrmMap[ng] = struct{}{}
		}

		for offset := -1 * lengthErrs; offset <= lengthErrs; offset++ {
			allowedErrs := abs(offset) - lengthErrs

			m := d.match(ngrms, l, offset)
			if m.IsEmpty() {
				continue
			}

			m.Iterate(func(id uint32) bool {
				doc, ok := d.docRaw(id)
				if !ok {
					return true
				}
				matches := 0
				for _, t := range doc.Terms {
					if _, ok := ngrmMap[t]; ok {
						matches++
					}
				}

				errCnt := abs(wordLen - matches - ngramSize + 1)
				if errCnt <= allowedErrs {
					result = append(result, Match{
						Value: doc.Value,
						Score: 1 / float64(errCnt) * math.Log1p(float64(d.counts[id])), // @todo
					})
				}

				return true
			})
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })

	if len(result) < maxCnt {
		return result
	}

	return result[0:maxCnt]
}

func (d *Dictionary) match(ngrm []string, wordLen int, offset int) *roaring.Bitmap {
	result := roaring.New()
	for i, ng := range ngrm {
		pos := i + offset
		if pos < 0 {
			continue
		}

		index := d.getIndex(wordLen, pos)
		m := index[ng]
		if m == nil {
			continue
		}
		result.Or(m)
	}

	return result
}

func abs(x int) int {
	if x < 0 {
		return -1 * x
	}
	return x
}
