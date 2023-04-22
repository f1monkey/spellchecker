package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_dictionary_id(t *testing.T) {
	dict, err := newDictionary(DefaultAlphabet, DefaultMaxErrors)
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
		dict, err := newDictionary(DefaultAlphabet, DefaultMaxErrors)
		require.NoError(t, err)

		id, err := dict.add("qwe")
		require.NoError(t, err)
		require.Equal(t, uint32(1), id)
		require.Equal(t, 1, dict.docs[id].Count)
		require.Equal(t, 1, len(dict.ids))
		require.Equal(t, uint32(2), dict.nextID)
		require.Len(t, dict.index, 1)

		id, err = dict.add("asd")
		require.NoError(t, err)
		require.Equal(t, uint32(2), id)
		require.Equal(t, 1, dict.docs[id].Count)
		require.Equal(t, 2, len(dict.ids))
		require.Equal(t, uint32(3), dict.nextID)
		require.Len(t, dict.index, 2)
	})
}

func Test_Dictionary_Inc(t *testing.T) {
	t.Run("must increase counter value", func(t *testing.T) {
		dict, err := newDictionary(DefaultAlphabet, DefaultMaxErrors)
		dict.docs[1] = Doc{}
		require.NoError(t, err)

		require.Equal(t, 0, dict.docs[1].Count)
		require.Equal(t, 0, dict.docs[2].Count)
		dict.inc(1)
		require.Equal(t, 1, dict.docs[1].Count)
		require.Equal(t, 0, dict.docs[2].Count)
	})
}
