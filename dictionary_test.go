package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_dictionary_id(t *testing.T) {
	dict := newDictionary()

	t.Run("must return 0 for unexisting word", func(t *testing.T) {
		id := dict.id("word")
		require.Equal(t, uint32(0), id)
	})

	t.Run("must return id for unexisting word", func(t *testing.T) {
		dict.IDs["word"] = 1
		id := dict.id("word")
		require.Equal(t, uint32(1), id)
	})
}

func Test_Dict_Add(t *testing.T) {
	t.Run("must add word to dictionary index", func(t *testing.T) {
		dict := newDictionary()
		id := dict.add("word", []Term{{Value: "word"}})
		require.Equal(t, uint32(1), id)
		require.Equal(t, 1, dict.Counts[id])
		require.Equal(t, 1, len(dict.IDs))
		require.Equal(t, uint32(2), dict.NextID)
		require.Equal(t, uint64(1), dict.Index["word"].GetCardinality())
		require.Nil(t, dict.Index["word2"])

		id = dict.add("word2", []Term{{Value: "word2"}})
		require.Equal(t, uint32(2), id)
		require.Equal(t, 1, dict.Counts[id])
		require.Equal(t, 2, len(dict.IDs))
		require.Equal(t, uint32(3), dict.NextID)
		require.Equal(t, uint64(1), dict.Index["word"].GetCardinality())
		require.Equal(t, uint64(1), dict.Index["word2"].GetCardinality())
	})
}

func Test_dictionary_inc(t *testing.T) {
	t.Run("must increase counter value", func(t *testing.T) {
		dict := newDictionary()
		require.Equal(t, 0, dict.Counts[1])
		require.Equal(t, 0, dict.Counts[2])
		dict.inc(1)
		require.Equal(t, 1, dict.Counts[1])
		require.Equal(t, 0, dict.Counts[2])
	})
}
