package spellchecker

import (
	"os"
	"path"
	"testing"

	"github.com/f1monkey/spellchecker/v4/internal/alphabet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSpellchecker(t *testing.T) {
	t.Parallel()

	s, err := New(&tokenizerMock{}, alphabet.EN)
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

		// "ban" and "bin" are the same edit away from "ben", so the higher weight must rank first.
		s.AddWeight(2, "ban")
		s.AddWeight(3, "ban")
		s.AddWeight(1, "bin")

		result := s.Suggest("ben", 2)
		require.Equal(t, []string{"ban", "bin"}, suggestionValues(result))
	})

	t.Run("adds remaining words if an earlier one already exists", func(t *testing.T) {
		t.Parallel()

		s, err := New(NewWhitespaceTokenizer(), EN)
		require.NoError(t, err)

		s.Add("ban")
		s.AddWeight(1, "ban", "bin")

		require.True(t, s.IsCorrect("ban"))
		require.True(t, s.IsCorrect("bin"))

		result := s.Suggest("ben", 2)
		require.Equal(t, []string{"ban", "bin"}, suggestionValues(result))
	})
}

func suggestionValues(result SuggestionResult) []string {
	values := make([]string, len(result.Suggestions))
	for i, match := range result.Suggestions {
		values[i] = match.Value
	}

	return values
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
		s.AddWeight(1, "dig's")

		require.True(t, s.IsCorrect("dog's"))
		require.True(t, s.IsCorrect("bone"))

		// "dog's" was added with a higher weight than "dig's", and both are one edit from "dug's".
		result := s.Suggest("dug's", 2)
		require.Equal(t, []string{"dog's", "dig's"}, suggestionValues(result))
	})
}

func Test_Spellchecker_Save(t *testing.T) {
	t.Parallel()

	m1 := newSampleSpellchecker(t)

	filePath := path.Join(t.TempDir(), "spellchecker.bin")
	file, err := os.Create(filePath)
	require.NoError(t, err)
	err = m1.Save(file)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	file, err = os.Open(filePath)
	require.NoError(t, err)

	m2, err := Load(file, NewWhitespaceTokenizer())
	require.NoError(t, err)

	require.Equal(t, m1.dict.ID("green"), m2.dict.ID("green"))

	// A word added after loading must get the same id, so the id sequence survived the round trip.
	m1.Add("brandnew")
	m2.Add("brandnew")
	require.Equal(t, m1.dict.ID("brandnew"), m2.dict.ID("brandnew"))
	require.NotEqual(t, m2.dict.ID("green"), m2.dict.ID("brandnew"))

	require.Equal(t, m1.Suggest("arang", 5), m2.Suggest("arang", 5))
}

type tokenizerMock struct{}

func (m *tokenizerMock) Tokenize(input string) []string {
	return []string{input}
}
