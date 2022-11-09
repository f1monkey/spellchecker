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

	ngrmMap := make(map[string]struct{})
	for _, ng := range ngrms {
		ngrmMap[ng] = struct{}{}
	}

	var result []Match

	wordLen := len([]rune(word))
	m := d.match(ngrms, wordLen)

	docs := make([]string, 0)

	m.Iterate(func(id uint32) bool {
		doc, ok := d.docRaw(id)
		if !ok {
			return true
		}
		docs = append(docs, doc.Value)
		matches := 0
		for _, t := range doc.Terms {
			if _, ok := ngrmMap[t]; ok {
				matches++
			}
		}

		errCnt := abs(wordLen - matches - ngramSize + 1)
		if errCnt <= maxErrors {
			result = append(result, Match{
				Value: doc.Value,
				Score: 1 / float64(errCnt) * math.Log1p(float64(d.counts[id])), // @todo
			})
		}

		return true
	})

	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })

	if len(result) < n {
		return result
	}

	return result[0:n]
}

func (d *Dictionary) match(ngrm []string, wordLen int) *roaring.Bitmap {
	result := roaring.New()
	index := d.getIndexByLen(wordLen)
	for _, ng := range ngrm {
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
