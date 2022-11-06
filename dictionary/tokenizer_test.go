package dictionary

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func Benchmark_makeTerms(b *testing.B) {
	for i := 0; i < b.N; i++ {
		makeTerms("qwerty")
	}
}

func Test_makeTerms(t *testing.T) {
	t.Run("must return nil for an empty word", func(t *testing.T) {
		require.Nil(t, makeTerms(""))
	})

	t.Run("must create a valid set of terms", func(t *testing.T) {
		expected := []Term{
			{Value: "t", Position: 0},
			{Value: "h", Position: 1},
			{Value: "e", Position: 2},
			{Value: "th", Position: 0},
			{Value: "he", Position: 1},
			{Value: "the", Position: 0},
		}

		require.Equal(t, expected, makeTerms("the"))
	})

	t.Run("must create max %d-length ngrams", func(t *testing.T) {
		result := makeTerms("qwertyuiop")
		for _, term := range result {
			require.LessOrEqual(t, utf8.RuneCountInString(term.Value), maxNGram)
		}
	})
}

func Benchmark_makeNGrams_6_3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		makeNGrams("qwerty", 3)
	}
}

func Benchmark_makeNGrams_5_3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		makeNGrams("qwert", 3)
	}
}

func Test_makeNGrams(t *testing.T) {
	t.Run("must panic if n < 1", func(t *testing.T) {
		require.Panics(t, func() {
			makeNGrams("word", 0)
		})
	})

	t.Run("must return nil if n > word len", func(t *testing.T) {
		require.Nil(t, makeNGrams("qwe", 5))
	})

	t.Run("must return word if n == word len", func(t *testing.T) {
		require.Equal(t, []string{"qwe"}, makeNGrams("qwe", 3))
	})

	t.Run("must return correct set of ngrams", func(t *testing.T) {
		t.Run("en", func(t *testing.T) {
			t.Run("len=4, n=3", func(t *testing.T) {
				require.Equal(t, []string{"wor", "ord"}, makeNGrams("word", 3))
			})
			t.Run("len=4, n=2", func(t *testing.T) {
				require.Equal(t, []string{"wo", "or", "rd"}, makeNGrams("word", 2))
			})
			t.Run("len=6, n=3", func(t *testing.T) {
				require.Equal(t, []string{"qwe", "wer", "ert", "rty"}, makeNGrams("qwerty", 3))
			})
		})

		t.Run("ru", func(t *testing.T) {
			t.Run("len=6, n=3", func(t *testing.T) {
				require.Equal(t, []string{"ябл", "бло", "лок", "око"}, makeNGrams("яблоко", 3))
			})
		})
	})
}
