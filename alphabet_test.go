package spellchecker

import (
	"testing"

	"github.com/f1monkey/bitmap"
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

func Test_alphabet_encode(t *testing.T) {
	t.Parallel()

	ab, err := newAlphabet("abcd")
	require.NoError(t, err)

	word := []rune("aab")
	result := ab.encode(word)
	require.Equal(t, bitmap.Bitmap32{3}, result)
}
