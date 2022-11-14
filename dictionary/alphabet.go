package dictionary

import "fmt"

const DefaultAlphabet = "abcdefghijklmnopqrstuvwxyz"

type alphabet map[rune]uint32

func newAlphabet(str string) (alphabet, error) {
	runes := []rune(str)
	if len(runes) > 63 {
		return nil, fmt.Errorf("alphabets longer than 63 are not supported yet")
	}

	result := make(alphabet)

	for i, s := range runes {
		if _, ok := result[s]; ok {
			return nil, fmt.Errorf("duplicate symbol %q at position %d", s, i)
		}
		result[s] = uint32(i)
	}

	return result, nil
}

func (a alphabet) encode(word []rune) uint64 {
	var result uint64

	for _, letter := range word {
		if index, ok := a[letter]; ok {
			result |= (1 << index)
		}
	}

	return result
}
