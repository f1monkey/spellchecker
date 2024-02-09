package spellchecker

import (
	"fmt"

	"github.com/f1monkey/bitmap"
)

const DefaultAlphabet = "abcdefghijklmnopqrstuvwxyz"

type alphabet map[rune]uint32

// newAlphabet create a new alphabet instance
func newAlphabet(str string) (alphabet, error) {
	runes := []rune(str)
	if len(runes) == 0 {
		return nil, fmt.Errorf("unable to use empty string as an alphabet")
	}

	result := make(alphabet, len(runes))
	for i, s := range runes {
		if _, ok := result[s]; ok {
			return nil, fmt.Errorf("duplicate symbol %q at position %d", s, i)
		}
		result[s] = uint32(i)
	}

	return result, nil
}

func (a alphabet) encode(word []rune) bitmap.Bitmap32 {
	var b bitmap.Bitmap32
	for _, letter := range word {
		if index, ok := a[letter]; ok {
			b.Set(index)
		}
	}

	return b
}

func (a alphabet) len() int {
	return len(a)
}
