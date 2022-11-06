package spellchecker

import (
	"errors"
	"os"
)

func loadFullSpellchecker() *Spellchecker {
	var s *Spellchecker
	ff, err := os.Open("data/spellchecker.bin")
	if errors.Is(err, os.ErrNotExist) {
		s = newFullSpellchecker()
		dst, err := os.Create("data/spellchecker.bin")
		if err != nil {
			panic(err)
		}

		err = s.Save(dst)
		if err != nil {
			panic(err)
		}
	} else {
		s, err = Load(ff)
		if err != nil {
			panic(err)
		}
	}

	return s
}

func newFullSpellchecker() *Spellchecker {
	f, err := os.Open("data/big.txt")
	if err != nil {
		panic(err)
	}

	s, err := New()
	if err != nil {
		panic(err)
	}

	err = s.AddFrom(f)
	if err != nil {
		panic(err)
	}

	return s
}

func newSampleSpellchecker() *Spellchecker {
	f, err := os.Open("data/sample.txt")
	if err != nil {
		panic(err)
	}

	s, err := New()
	if err != nil {
		panic(err)
	}
	err = s.AddFrom(f)
	if err != nil {
		panic(err)
	}

	return s
}
