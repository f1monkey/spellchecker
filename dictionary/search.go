package dictionary

import "github.com/RoaringBitmap/roaring"

type Hit struct {
	Value string
	Score float64
}

// Find find top N hits by word
func (d *Dictionary) Find(word string, n int) []Hit {
	m := d.match(makeTerms(word))
	if m.IsEmpty() {
		return nil
	}

	// @todo calculate scores and return top hits
	result := make([]Hit, 0, m.GetCardinality())
	m.Iterate(func(x uint32) bool {
		doc, ok := d.docs[x]
		if !ok {
			return true
		}

		result = append(result, Hit{
			Value: doc.Value,
		})

		return true
	})

	return result
}

func (d *Dictionary) match(terms []Term) *roaring.Bitmap {
	result := roaring.New()
	for _, t := range terms {
		m := d.index[t.Value]
		if m == nil {
			continue
		}
		result.Or(m)
	}

	return result
}
