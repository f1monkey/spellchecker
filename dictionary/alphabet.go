package dictionary

import "fmt"

var DefaultAlphabet = AlphabetConfig{
	Letters: "abcdefghijklmnopqrstuvwxyz",
	Length:  26,
}

type alphabet map[rune]uint32

// newAlphabet create a new alphabet instance
func newAlphabet(str string, length int) (alphabet, error) {
	runes := []rune(str)
	if len(runes) == 0 {
		return nil, fmt.Errorf("unable to use empty string as alphabet")
	}

	if length > 63 {
		return nil, fmt.Errorf("alphabets longer than 63 are not supported yet")
	}

	result := make(alphabet)

	for i, s := range []rune(str) {
		index := i % length
		if _, ok := result[s]; ok {
			return nil, fmt.Errorf("duplicate symbol %q at position %d", s, i)
		}
		result[s] = uint32(index)
	}

	return result, nil
}

func (a alphabet) encode(word []rune) bitmap {
	var b bitmap
	for _, letter := range word {
		if index, ok := a[letter]; ok {
			b.or(index)
		}
	}

	return b
}

func (a alphabet) len() int {
	return len(a)
}
