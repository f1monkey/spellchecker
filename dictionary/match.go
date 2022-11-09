package dictionary

import (
	"github.com/RoaringBitmap/roaring"
	"github.com/cyradin/ngrams"
)

func (d *Dictionary) Match(word string) (*roaring.Bitmap, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	result := roaring.New()

	ngrm, err := ngrams.From(word, 3)
	if err != nil {
		return nil, err
	}

	index := d.getIndex(word)

	for _, ng := range ngrm {
		m := index[ng]
		if m == nil {
			continue
		}
		result.Or(m)
	}

	return result, nil
}
