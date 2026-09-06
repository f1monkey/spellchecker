package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSpellchecker(t *testing.T) {
	t.Parallel()

	s, err := New(&tokenizerMock{}, EN)
	require.NoError(t, err)
	require.NotNil(t, s.dict)
}

func Test_Spellchecker_IsCorrect(t *testing.T) {
	t.Parallel()

	s := newSampleSpellchecker(t)

	assert.True(t, s.IsCorrect("orange"))
	assert.False(t, s.IsCorrect("car"))
}

func Test_Spellchecker_Add(t *testing.T) {
	t.Parallel()

	t.Run("adds each argument as a whole word", func(t *testing.T) {
		t.Parallel()

		s, err := New(NewWhitespaceTokenizer(), EN)
		require.NoError(t, err)

		s.Add("hello world", "foo")

		require.True(t, s.IsCorrect("hello world"))
		require.True(t, s.IsCorrect("foo"))
		require.False(t, s.IsCorrect("hello"))
	})

	t.Run("skips empty strings", func(t *testing.T) {
		t.Parallel()

		s, err := New(NewWhitespaceTokenizer(), EN)
		require.NoError(t, err)

		s.Add("", "bar", "")

		require.True(t, s.IsCorrect("bar"))
		require.False(t, s.IsCorrect(""))
	})
}

func Test_Spellchecker_AddWeight(t *testing.T) {
	t.Parallel()

	t.Run("increments count for an existing word", func(t *testing.T) {
		t.Parallel()

		s, err := New(NewWhitespaceTokenizer(), EN)
		require.NoError(t, err)

		s.AddWeight(2, "tea")
		s.AddWeight(3, "tea")

		id := s.dict.id("tea")
		require.Equal(t, uint(5), s.dict.counts[id])
	})

	t.Run("adds remaining words if an earlier one already exists", func(t *testing.T) {
		t.Parallel()

		s, err := New(NewWhitespaceTokenizer(), EN)
		require.NoError(t, err)

		s.Add("tea")
		s.AddWeight(1, "tea", "coffee")

		require.True(t, s.IsCorrect("tea"))
		require.True(t, s.IsCorrect("coffee"))
		require.Equal(t, uint(2), s.dict.counts[s.dict.id("tea")])
		require.Equal(t, uint(1), s.dict.counts[s.dict.id("coffee")])
	})
}

func Test_Spellchecker_AddPhrases(t *testing.T) {
	t.Parallel()

	t.Run("tokenizes phrases before adding", func(t *testing.T) {
		t.Parallel()

		s, err := New(NewWhitespaceTokenizer(), EN)
		require.NoError(t, err)

		s.AddPhrases("green tea", "black coffee")

		require.True(t, s.IsCorrect("green"))
		require.True(t, s.IsCorrect("tea"))
		require.True(t, s.IsCorrect("black"))
		require.True(t, s.IsCorrect("coffee"))
		require.False(t, s.IsCorrect("green tea"))
	})

	t.Run("applies weight to every token", func(t *testing.T) {
		t.Parallel()

		s, err := New(NewStandardTokenizer(), EN)
		require.NoError(t, err)

		s.AddPhraseWeight(4, "dog's bone")

		require.Equal(t, uint(4), s.dict.counts[s.dict.id("dog's")])
		require.Equal(t, uint(4), s.dict.counts[s.dict.id("bone")])
	})
}

type tokenizerMock struct{}

func (m *tokenizerMock) Tokenize(input string) []string {
	return []string{input}
}
