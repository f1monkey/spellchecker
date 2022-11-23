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
		dict, err := newDictionary(DefaultAlphabet)
		dict.docs[1] = Doc{}
		require.NoError(t, err)

		require.Equal(t, 0, dict.docs[1].Count)
		require.Equal(t, 0, dict.docs[2].Count)
		dict.inc(1)
		require.Equal(t, 1, dict.docs[1].Count)
		require.Equal(t, 0, dict.docs[2].Count)
	})
}

// func Benchmark_match(b *testing.B) {
// 	dict := New()
// 	dict.Add("orange")
// 	dict.Add("range")

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		dict.match([]string{"ran", "nge"}, 5, 0)
// 	}
// }

// func Test_Match(t *testing.T) {
// 	dict := New()
// 	dict.Add("orange")
// 	dict.Add("ranger")

// 	t.Run("must return empty bitmap if nothing found", func(t *testing.T) {
// 		m := dict.match([]string{"qwe"}, 3, 0)
// 		require.NotNil(t, m)
// 		require.True(t, m.IsEmpty())
// 	})
// 	t.Run("must be able to find a word without offset", func(t *testing.T) {
// 		m := dict.match([]string{"ora", "ran", "ang", "nge"}, 6, 0)
// 		require.False(t, m.IsEmpty())
// 		require.Equal(t, uint64(1), m.GetCardinality())

// 		doc, ok := dict.docRaw(m.ToArray()[0])
// 		require.True(t, ok)
// 		require.Equal(t, "orange", doc.Word)
// 	})
// 	t.Run("must be able to find a word with offset", func(t *testing.T) {
// 		m := dict.match([]string{"ora", "ran", "ang", "nge"}, 6, -1)
// 		require.False(t, m.IsEmpty())
// 		require.Equal(t, uint64(1), m.GetCardinality())

// 		doc, ok := dict.docRaw(m.ToArray()[0])
// 		require.True(t, ok)
// 		require.Equal(t, "ranger", doc.Word)
// 	})
// }
