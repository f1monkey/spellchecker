package alphabet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()

	t.Run("must not allow an empty string to be the alphabet", func(t *testing.T) {
		t.Parallel()

		result, err := New("")
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("must create a valid map from the string", func(t *testing.T) {
		t.Parallel()

		result, err := New("abc")
		require.NoError(t, err)
		require.Equal(t, result, Alphabet{'a': 0, 'b': 1, 'c': 2})
	})

	t.Run("must not allow duplicate symbols in alphabet", func(t *testing.T) {
		t.Parallel()

		result, err := New("abb")
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func Test_Alphabet_key(t *testing.T) {
	t.Parallel()

	ab, err := New("abcd")
	require.NoError(t, err)

	zobrist := ab.Zobrist()

	t.Run("must depend only on the set of letters", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, ab.Key([]rune("ab"), zobrist), ab.Key([]rune("aab"), zobrist))
		require.Equal(t, ab.Key([]rune("ab"), zobrist), ab.Key([]rune("bba"), zobrist))
		require.NotEqual(t, ab.Key([]rune("ab"), zobrist), ab.Key([]rune("abc"), zobrist))
	})

	t.Run("must ignore symbols outside the alphabet", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, ab.Key([]rune("ab"), zobrist), ab.Key([]rune("a-b!"), zobrist))
		require.Equal(t, uint64(0), ab.Key([]rune("xyz"), zobrist))
	})

	t.Run("must flip a letter with xor", func(t *testing.T) {
		t.Parallel()

		key := ab.Key([]rune("ab"), zobrist)
		require.Equal(t, ab.Key([]rune("abc"), zobrist), key^zobrist[2])
		require.Equal(t, ab.Key([]rune("a"), zobrist), key^zobrist[1])
	})
}

func Test_Alphabet_zobrist(t *testing.T) {
	t.Parallel()

	ab, err := New(EN, RU, Numbers)
	require.NoError(t, err)

	zobrist := ab.Zobrist()
	require.Len(t, zobrist, len(ab))
	require.Equal(t, zobrist, ab.Zobrist())

	seen := make(map[uint64]struct{}, len(zobrist))
	for _, v := range zobrist {
		seen[v] = struct{}{}
	}

	require.Len(t, seen, len(zobrist))
}
