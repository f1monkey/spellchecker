package dictmemory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Dict_ID(t *testing.T) {
	dict := NewDict()

	t.Run("must return 0 for unexisting word", func(t *testing.T) {
		id, err := dict.ID("word")
		require.NoError(t, err)
		require.Equal(t, uint32(0), id)
	})

	t.Run("must return id for unexisting word", func(t *testing.T) {
		dict.IDs["word"] = 1
		id, err := dict.ID("word")
		require.NoError(t, err)
		require.Equal(t, uint32(1), id)
	})
}

func Test_Dict_Add(t *testing.T) {
	dict := NewDict()

	t.Run("must add unexisting word", func(t *testing.T) {
		id, err := dict.Add("word")
		require.NoError(t, err)
		require.Equal(t, uint32(1), id)
		require.Equal(t, 1, dict.Counts[id])
		require.Equal(t, 1, len(dict.IDs))
		require.Equal(t, uint32(2), dict.NextID)

		id, err = dict.Add("word2")
		require.NoError(t, err)
		require.Equal(t, uint32(2), id)
		require.Equal(t, 1, dict.Counts[id])
		require.Equal(t, 2, len(dict.IDs))
		require.Equal(t, uint32(3), dict.NextID)
	})

	t.Run("must inc counter for existing word", func(t *testing.T) {
		id, err := dict.Add("word")
		require.NoError(t, err)

		require.Equal(t, uint32(1), id)
		require.Equal(t, 2, dict.Counts[id])
		require.Equal(t, 2, len(dict.IDs))
		require.Equal(t, uint32(3), dict.NextID)
	})
}
