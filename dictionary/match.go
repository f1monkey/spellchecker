package dictionary

import (
	"github.com/RoaringBitmap/roaring"
	"github.com/cyradin/spellchecker/ngram"
)

func (d *Dictionary) Match(terms []ngram.NGram) *roaring.Bitmap {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

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
