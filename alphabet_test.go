package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newAlphabet(t *testing.T) {
	t.Parallel()

	t.Run("must not allow an empty string to be the alphabet", func(t *testing.T) {
		t.Parallel()

		result, err := newAlphabet("")
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("must create a valid map from the string", func(t *testing.T) {
		t.Parallel()

		result, err := newAlphabet("abc")
		require.NoError(t, err)
		require.Equal(t, result, alphabet{'a': 0, 'b': 1, 'c': 2})
	})

	t.Run("must not allow duplicate symbols in alphabet", func(t *testing.T) {
		t.Parallel()

		result, err := newAlphabet("abb")
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func Test_alphabet_key(t *testing.T) {
	t.Parallel()

	ab, err := newAlphabet("abcd")
	require.NoError(t, err)

	zobrist := ab.zobrist()

	t.Run("must depend only on the set of letters", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, ab.key([]rune("ab"), zobrist), ab.key([]rune("aab"), zobrist))
		require.Equal(t, ab.key([]rune("ab"), zobrist), ab.key([]rune("bba"), zobrist))
		require.NotEqual(t, ab.key([]rune("ab"), zobrist), ab.key([]rune("abc"), zobrist))
	})

	t.Run("must ignore symbols outside the alphabet", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, ab.key([]rune("ab"), zobrist), ab.key([]rune("a-b!"), zobrist))
		require.Equal(t, uint64(0), ab.key([]rune("xyz"), zobrist))
	})

	t.Run("must flip a letter with xor", func(t *testing.T) {
		t.Parallel()

		key := ab.key([]rune("ab"), zobrist)
		require.Equal(t, ab.key([]rune("abc"), zobrist), key^zobrist[2])
		require.Equal(t, ab.key([]rune("a"), zobrist), key^zobrist[1])
	})
}

func Test_alphabet_zobrist(t *testing.T) {
	t.Parallel()

	ab, err := newAlphabet(EN, RU, Numbers)
	require.NoError(t, err)

	zobrist := ab.zobrist()
	require.Len(t, zobrist, len(ab))
	require.Equal(t, zobrist, ab.zobrist())

	seen := make(map[uint64]struct{}, len(zobrist))
	for _, v := range zobrist {
		seen[v] = struct{}{}
	}

	require.Len(t, seen, len(zobrist))
}
