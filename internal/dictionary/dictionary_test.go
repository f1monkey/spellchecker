package dictionary

import (
	"testing"

	"github.com/f1monkey/spellchecker/v4/internal/alphabet"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()

	t.Run("must reject an empty alphabet", func(t *testing.T) {
		t.Parallel()

		dict, err := New("")
		require.Error(t, err)
		require.Nil(t, dict)
	})
}

func Test_Dictionary_Has(t *testing.T) {
	t.Parallel()

	dict, err := New(alphabet.EN)
	require.NoError(t, err)

	require.False(t, dict.Has("word"))

	dict.Add("word", 1)
	require.True(t, dict.Has("word"))
}

func Test_Dictionary_ID(t *testing.T) {
	t.Parallel()

	t.Run("must return 0 for unexisting word", func(t *testing.T) {
		t.Parallel()

		dict, err := New(alphabet.EN)
		require.NoError(t, err)

		id := dict.ID("word")
		require.Equal(t, uint32(0), id)
	})

	t.Run("must return id for existing word", func(t *testing.T) {
		t.Parallel()

		dict, err := New(alphabet.EN)
		require.NoError(t, err)

		dict.ids["word"] = 1
		id := dict.ID("word")
		require.Equal(t, uint32(1), id)
	})
}

func Test_Dictionary_Add(t *testing.T) {
	t.Parallel()

	t.Run("must add word to dictionary index", func(t *testing.T) {
		t.Parallel()

		dict, err := New(alphabet.EN)
		require.NoError(t, err)

		id := dict.Add("qwe", 1)
		require.Equal(t, uint32(1), id)
		require.Equal(t, uint(1), dict.counts[id])
		require.Equal(t, []rune("qwe"), dict.words[id])
		require.Equal(t, 1, len(dict.ids))
		require.Len(t, dict.index, 1)

		id = dict.Add("asd", 2)
		require.Equal(t, uint32(2), id)
		require.Equal(t, uint(2), dict.counts[id])
		require.Equal(t, []rune("asd"), dict.words[id])
		require.Equal(t, 2, len(dict.ids))
		require.Len(t, dict.index, 2)

		require.Equal(t, uint32(3), dict.nextID())
	})
}

func Test_Dictionary_Inc(t *testing.T) {
	t.Parallel()

	t.Run("must increase counter value", func(t *testing.T) {
		t.Parallel()

		dict, err := New(alphabet.EN)
		require.NoError(t, err)

		dict.counts[1] = 0
		require.Equal(t, uint(0), dict.counts[1])
		require.Equal(t, uint(0), dict.counts[2])

		dict.Inc(1, 100)

		require.Equal(t, uint(100), dict.counts[1])
		require.Equal(t, uint(0), dict.counts[2])
	})

	t.Run("must ignore an unknown id", func(t *testing.T) {
		t.Parallel()

		dict, err := New(alphabet.EN)
		require.NoError(t, err)

		dict.Inc(42, 5)
		require.NotContains(t, dict.counts, uint32(42))
	})
}

func Test_Dictionary_Find(t *testing.T) {
	t.Parallel()

	t.Run("must return nothing when max errors is not positive", func(t *testing.T) {
		t.Parallel()

		dict := mustDictionary(t, "orange")

		require.Nil(t, dict.Find("oragne", 5, 0, acceptAll))
	})

	t.Run("must stop after the same letter set", func(t *testing.T) {
		t.Parallel()

		// "oragne" uses the same letters as "orange", so the search returns before
		// flipping bits. "green" is within two flips and would show up otherwise.
		dict := mustDictionary(t, "orange", "green")

		require.Equal(t, []Match{{Value: "orange", Score: 1}}, dict.Find("oragne", 5, 2, acceptAll))
	})

	t.Run("must find a word that differs by up to max errors", func(t *testing.T) {
		t.Parallel()

		dict := mustDictionary(t, "problem")

		require.Equal(t, []Match{{Value: "problem", Score: 1}}, dict.Find("problam", 5, 2, acceptAll))
		require.Empty(t, dict.Find("problam", 5, 1, acceptAll))
	})

	t.Run("must rank by score and keep only the requested number", func(t *testing.T) {
		t.Parallel()

		dict, err := New(alphabet.EN)
		require.NoError(t, err)

		dict.Add("ab", 1)
		dict.Add("ba", 3)

		require.Equal(t, []Match{{Value: "ba", Score: 3}}, dict.Find("ab", 1, 2, acceptAll))
	})

	t.Run("must skip candidates rejected by the filter", func(t *testing.T) {
		t.Parallel()

		dict := mustDictionary(t, "problem")
		reject := func([]rune, []rune, uint, int) (float64, bool) {
			return 0, false
		}

		require.Empty(t, dict.Find("problam", 5, 2, reject))
	})

	t.Run("must skip an index entry whose word is missing", func(t *testing.T) {
		t.Parallel()

		dict := mustDictionary(t, "orange")
		delete(dict.words, dict.ID("orange"))

		require.Empty(t, dict.Find("oragne", 5, 2, acceptAll))
	})
}

func Test_Dictionary_MarshalBinary(t *testing.T) {
	t.Parallel()

	t.Run("must restore words, counts, index and the id sequence", func(t *testing.T) {
		t.Parallel()

		dict := mustDictionary(t, "orange", "green")

		data, err := dict.MarshalBinary()
		require.NoError(t, err)

		loaded := &Dictionary{}
		require.NoError(t, loaded.UnmarshalBinary(data))

		require.Equal(t, dict.ids, loaded.ids)
		require.Equal(t, dict.counts, loaded.counts)
		require.Equal(t, dict.words, loaded.words)
		require.Equal(t, dict.index, loaded.index)
		require.Equal(t, dict.zobrist, loaded.zobrist)

		require.Equal(t, dict.nextID(), loaded.nextID())
		require.Equal(t, dict.Find("oragne", 5, 2, acceptAll), loaded.Find("oragne", 5, 2, acceptAll))
	})

	t.Run("must skip ids that have no word while rebuilding the index", func(t *testing.T) {
		t.Parallel()

		dict := mustDictionary(t, "one", "two", "three")
		delete(dict.words, dict.ID("two"))

		data, err := dict.MarshalBinary()
		require.NoError(t, err)

		loaded := &Dictionary{}
		require.NoError(t, loaded.UnmarshalBinary(data))

		require.True(t, loaded.Has("one"))
		require.True(t, loaded.Has("three"))
		require.NotContains(t, loaded.words, dict.ID("two"))
		require.Equal(t, dict.Find("one", 5, 2, acceptAll), loaded.Find("one", 5, 2, acceptAll))
	})

	t.Run("must return an error for malformed data", func(t *testing.T) {
		t.Parallel()

		dict := &Dictionary{}
		require.Error(t, dict.UnmarshalBinary([]byte("not a dictionary")))
	})
}

func mustDictionary(t *testing.T, words ...string) *Dictionary {
	t.Helper()

	dict, err := New(alphabet.EN)
	require.NoError(t, err)

	for _, word := range words {
		dict.Add(word, 1)
	}

	return dict
}

func acceptAll(_, _ []rune, count uint, _ int) (float64, bool) {
	return float64(count), true
}
