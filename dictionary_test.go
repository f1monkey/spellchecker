package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_dictionary_id(t *testing.T) {
	dict, err := newDictionary(DefaultAlphabet)
	require.NoError(t, err)

	t.Run("must return 0 for unexisting word", func(t *testing.T) {
		id := dict.id("word")
		require.Equal(t, uint32(0), id)
	})

	t.Run("must return id for unexisting word", func(t *testing.T) {
		dict.ids["word"] = 1
		id := dict.id("word")
		require.Equal(t, uint32(1), id)
	})
}

func Test_dictionary_add(t *testing.T) {
	t.Run("must add word to dictionary index", func(t *testing.T) {
		dict, err := newDictionary(DefaultAlphabet)
		require.NoError(t, err)

		id, err := dict.add("qwe", 1)
		require.NoError(t, err)
		require.Equal(t, uint32(1), id)
		require.Equal(t, uint(1), dict.counts[id])
		require.Equal(t, []rune("qwe"), dict.words[id])
		require.Equal(t, 1, len(dict.ids))
		require.Len(t, dict.index, 1)

		id, err = dict.add("asd", 2)
		require.NoError(t, err)
		require.Equal(t, uint32(2), id)
		require.Equal(t, uint(2), dict.counts[id])
		require.Equal(t, []rune("asd"), dict.words[id])
		require.Equal(t, 2, len(dict.ids))
		require.Len(t, dict.index, 2)

		require.Equal(t, uint32(3), dict.nextID())
	})
}

func Test_Dictionary_Inc(t *testing.T) {
	t.Run("must increase counter value", func(t *testing.T) {
		dict, err := newDictionary(DefaultAlphabet)
		dict.counts[1] = 0
		require.NoError(t, err)

		require.Equal(t, uint(0), dict.counts[1])
		require.Equal(t, uint(0), dict.counts[2])
		dict.inc(1, 100)
		require.Equal(t, uint(100), dict.counts[1])
		require.Equal(t, uint(0), dict.counts[2])
	})
}
