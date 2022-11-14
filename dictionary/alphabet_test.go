package dictionary

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newAlphabet(t *testing.T) {
	t.Run("must create a valid map from the string", func(t *testing.T) {
		result, err := newAlphabet("abc")
		require.NoError(t, err)
		require.Equal(t, result, alphabet{'a': 0, 'b': 1, 'c': 2})
	})

	t.Run("must return error if alphabet length is greater than max", func(t *testing.T) {
		result, err := newAlphabet("abcdefghijklmnopqrstuvwxyzабвгдеёжзийклмнопрстуфхцчшщъыьэюя01234")
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("must not allow duplicate symbols in alphabet", func(t *testing.T) {
		result, err := newAlphabet("abb")
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func Test_alphabet_encode(t *testing.T) {
	ab, err := newAlphabet(DefaultAlphabet)
	require.NoError(t, err)

	word := []rune("word")
	result := ab.encode(word)
	require.Equal(t, uint64(4341768), result)
}
