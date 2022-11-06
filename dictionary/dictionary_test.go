package dictionary

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Dictionary_ID(t *testing.T) {
	dict := New()

	t.Run("must return 0 for unexisting word", func(t *testing.T) {
		id := dict.ID("word")
		require.Equal(t, uint32(0), id)
	})

	t.Run("must return id for unexisting word", func(t *testing.T) {
		dict.IDs["word"] = 1
		id := dict.ID("word")
		require.Equal(t, uint32(1), id)
	})
}

func Test_Dictionary_Add(t *testing.T) {
	t.Run("must add word to dictionary index", func(t *testing.T) {
		dict := New()
		id := dict.Add("qwe")
		require.Equal(t, uint32(1), id)
		require.Equal(t, 1, dict.Counts[id])
		require.Equal(t, 1, len(dict.IDs))
		require.Equal(t, uint32(2), dict.NextID)
		require.Equal(t, uint64(1), dict.Index["qwe"].GetCardinality())
		require.Nil(t, dict.Index["asd"])

		id = dict.Add("asd")
		require.Equal(t, uint32(2), id)
		require.Equal(t, 1, dict.Counts[id])
		require.Equal(t, 2, len(dict.IDs))
		require.Equal(t, uint32(3), dict.NextID)
		require.Equal(t, uint64(1), dict.Index["qwe"].GetCardinality())
		require.Equal(t, uint64(1), dict.Index["asd"].GetCardinality())
	})
}

func Test_Dictionary_Inc(t *testing.T) {
	t.Run("must increase counter value", func(t *testing.T) {
		dict := New()
		require.Equal(t, 0, dict.Counts[1])
		require.Equal(t, 0, dict.Counts[2])
		dict.Inc(1)
		require.Equal(t, 1, dict.Counts[1])
		require.Equal(t, 0, dict.Counts[2])
	})
}
