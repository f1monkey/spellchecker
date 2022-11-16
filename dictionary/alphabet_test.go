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
	ab, err := newAlphabet("abcd")
	require.NoError(t, err)

	word := []rune("aab")
	result := ab.encode(word)
	require.Equal(t, bitmap(3), result)
}

func Benchmark_alphabet_orAll(b *testing.B) {
	ab, err := newAlphabet(DefaultAlphabet)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		ab.orAll(bitmap(0))
	}
}

func Test_alphabet_orAll(t *testing.T) {
	ab, err := newAlphabet("abcd")
	require.NoError(t, err)

	var b bitmap
	b.or(1)

	result := ab.orAll(b)
	require.Equal(t, []bitmap{3, 2, 6, 10}, result)
}