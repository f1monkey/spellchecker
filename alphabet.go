package spellchecker

import (
	"fmt"
	"slices"
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

// zobrist returns a pseudo-random 64-bit value for every alphabet symbol.
// The values depend only on the symbol position, so they are stable across runs.
func (a alphabet) zobrist() []uint64 {
	result := make([]uint64, len(a))
	for i := range result {
		result[i] = splitmix64(uint64(i))
	}

	return result
}

// key returns the Zobrist hash of the set of alphabet symbols used in the word.
// Repeated symbols and symbols outside the alphabet do not affect the key.
func (a alphabet) key(word []rune, zobrist []uint64) uint64 {
	var result uint64

	for i, letter := range word {
		index, ok := a[letter]
		if !ok || slices.Contains(word[:i], letter) {
			continue
		}

		result ^= zobrist[index]
	}

	return result
}

// splitmix64 is the SplitMix64 hash (Steele, Lea, Flood, 2014): it maps
// sequential inputs to well-mixed, independent-looking 64-bit values.
// Every step is invertible, so distinct inputs always give distinct outputs.
func splitmix64(x uint64) uint64 {
	const (
		// 2^64 / golden ratio; odd, so the addition is a bijection.
		// It also keeps splitmix64(0) from being 0.
		golden = 0x9e3779b97f4a7c15

		// David Stafford's "Mix13" finalizer parameters (a tuned variant of
		// MurmurHash3 fmix64), chosen empirically for the best avalanche.
		mul1   = 0xbf58476d1ce4e5b9
		mul2   = 0x94d049bb133111eb
		shift1 = 30
		shift2 = 27
		shift3 = 31
	)

	x += golden
	x = (x ^ (x >> shift1)) * mul1
	x = (x ^ (x >> shift2)) * mul2

	return x ^ (x >> shift3)
}
