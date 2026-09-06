package spellchecker

import (
	"fmt"

	"github.com/f1monkey/bitmap"
)

type Alphabet = string

const (
	EN      Alphabet = "abcdefghijklmnopqrstuvwxyz"
	RU      Alphabet = "абвгдеёжзийклмнопрстуфхцчшщъыьэюя"
	Numbers Alphabet = "1234567890"
)

const DefaultAlphabet = EN

type alphabet map[rune]uint32

// newAlphabet create a new alphabet instance
func newAlphabet(strings ...Alphabet) (alphabet, error) {
	result := make(alphabet)

	var runes []rune

	for _, str := range strings {
		if len(str) == 0 {
			return nil, fmt.Errorf("unable to use empty string as an alphabet")
		}

		runes = append(runes, []rune(str)...)
	}

	for i, r := range runes {
		if _, ok := result[r]; ok {
			return nil, fmt.Errorf("duplicate symbol %q at position %d", r, i)
		}

		result[r] = uint32(i)
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
