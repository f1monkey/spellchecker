package ngram

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func Benchmark_MakeAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MakeAll("qwerty", 5)
	}
}

func Test_MakeAll(t *testing.T) {
	t.Run("must panic if max < 1", func(t *testing.T) {
		require.Panics(t, func() {
			MakeAll("word", 0)
		})
	})

	t.Run("must return nil for an empty word", func(t *testing.T) {
		require.Nil(t, MakeAll("", 5))
	})

	t.Run("must create a valid set of terms", func(t *testing.T) {
		expected := []NGram{
			{Value: "t", Position: 0},
			{Value: "h", Position: 1},
			{Value: "e", Position: 2},
			{Value: "th", Position: 0},
			{Value: "he", Position: 1},
			{Value: "the", Position: 0},
		}

		require.Equal(t, expected, MakeAll("the", 5))
	})

	t.Run("must create ngrams with length <= max", func(t *testing.T) {
		result := MakeAll("qwertyuiop", 5)
		for _, term := range result {
			require.LessOrEqual(t, utf8.RuneCountInString(term.Value), 5)
		}
	})
}

func Benchmark_Make_6_3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Make("qwerty", 3)
	}
}

func Benchmark_Make_5_3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Make("qwert", 3)
	}
}

func Test_Make(t *testing.T) {
	t.Run("must panic if n < 1", func(t *testing.T) {
		require.Panics(t, func() {
			Make("word", 0)
		})
	})

	t.Run("must return nil if n > word len", func(t *testing.T) {
		require.Nil(t, Make("qwe", 5))
	})

	t.Run("must return word if n == word len", func(t *testing.T) {
		require.Equal(t, []NGram{{Value: "qwe"}}, Make("qwe", 3))
	})

	t.Run("must return correct set of ngrams", func(t *testing.T) {
		t.Run("en", func(t *testing.T) {
			t.Run("len=4, n=3", func(t *testing.T) {
				require.Equal(t, []NGram{
					{Value: "wor", Position: 0},
					{Value: "ord", Position: 1}},
					Make("word", 3),
				)
			})
			t.Run("len=4, n=2", func(t *testing.T) {
				require.Equal(t, []NGram{
					{Value: "wo", Position: 0},
					{Value: "or", Position: 1},
					{Value: "rd", Position: 2},
				}, Make("word", 2))
			})
			t.Run("len=6, n=3", func(t *testing.T) {
				require.Equal(t, []NGram{
					{Value: "qwe", Position: 0},
					{Value: "wer", Position: 1},
					{Value: "ert", Position: 2},
					{Value: "rty", Position: 3},
				}, Make("qwerty", 3))
			})
		})

		t.Run("ru", func(t *testing.T) {
			t.Run("len=6, n=3", func(t *testing.T) {
				require.Equal(t, []NGram{
					{Value: "ябл", Position: 0},
					{Value: "бло", Position: 1},
					{Value: "лок", Position: 2},
					{Value: "око", Position: 3},
				}, Make("яблоко", 3))
			})
		})
	})
}
