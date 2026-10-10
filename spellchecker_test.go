package spellchecker

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSpellchecker(t *testing.T) {
	t.Parallel()

	s, err := New(EN)
	require.NoError(t, err)
	require.NotNil(t, s.dict)
}

func Test_Mistake_String(t *testing.T) {
	t.Parallel()

	require.Equal(t, "none", NoMistake.String())
	require.Equal(t, "typo", MistakeTypo.String())
	require.Equal(t, "typo|layout", (MistakeLayout | MistakeTypo).String())
	require.Equal(t, "unknown_word|extra_space", (MistakeUnknownWord | MistakeExtraSpace).String())
}

func Test_SuggestionResult_IsCorrect(t *testing.T) {
	t.Parallel()

	require.True(t, SuggestionResult{}.IsCorrect())
	require.False(t, SuggestionResult{Mistakes: MistakeTypo}.IsCorrect())
	require.False(t, SuggestionResult{Mistakes: MistakeUnknownWord}.IsCorrect())
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

		s, err := New(EN)
		require.NoError(t, err)

		s.Add("hello world", "foo")

		require.True(t, s.IsCorrect("hello world"))
		require.True(t, s.IsCorrect("foo"))
		require.False(t, s.IsCorrect("hello"))
	})

	t.Run("skips empty strings", func(t *testing.T) {
		t.Parallel()

		s, err := New(EN)
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

		s, err := New(EN)
		require.NoError(t, err)

		// "ban" and "bin" are the same edit away from "ben", so the higher weight must rank first.
		s.AddWeight(2, "ban")
		s.AddWeight(3, "ban")
		s.AddWeight(1, "bin")

		result := s.Suggest("ben", WithMaxSuggestions(2))
		require.Equal(t, []string{"ban", "bin"}, suggestionValues(result))
	})

	t.Run("adds remaining words if an earlier one already exists", func(t *testing.T) {
		t.Parallel()

		s, err := New(EN)
		require.NoError(t, err)

		s.Add("ban")
		s.AddWeight(1, "ban", "bin")

		require.True(t, s.IsCorrect("ban"))
		require.True(t, s.IsCorrect("bin"))

		result := s.Suggest("ben", WithMaxSuggestions(2))
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

func Test_Spellchecker_Save(t *testing.T) {
	t.Parallel()

	m1 := newSampleSpellchecker(t)

	filePath := path.Join(t.TempDir(), "bin")
	file, err := os.Create(filePath)
	require.NoError(t, err)
	err = m1.Save(file)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	file, err = os.Open(filePath)
	require.NoError(t, err)

	m2, err := Load(file)
	require.NoError(t, err)

	require.Equal(t, m1.dict.ID("green"), m2.dict.ID("green"))

	// A word added after loading must get the same id, so the id sequence survived the round trip.
	m1.Add("brandnew")
	m2.Add("brandnew")
	require.Equal(t, m1.dict.ID("brandnew"), m2.dict.ID("brandnew"))
	require.NotEqual(t, m2.dict.ID("green"), m2.dict.ID("brandnew"))

	require.Equal(t, m1.Suggest("arang", WithMaxSuggestions(5)), m2.Suggest("arang", WithMaxSuggestions(5)))
}

func Test_Spellchecker_Suggest(t *testing.T) {
	t.Parallel()

	t.Run("fix", func(t *testing.T) {
		t.Parallel()

		s := newSampleSpellchecker(t)
		result := s.Suggest("arang", WithMaxSuggestions(5))
		require.Equal(t, SuggestionResult{
			Mistakes: MistakeTypo,
			Suggestions: []Suggestion{
				{Value: "orange", Score: 0.2772588722239781},
				{Value: "range", Score: 0.13862943611198905},
			},
		}, result)
	})

	t.Run("custom max errors", func(t *testing.T) {
		t.Parallel()

		s := newSampleSpellchecker(t)
		result := s.Suggest("rang", WithMaxSuggestions(5), WithMaxErrors(1))
		require.Equal(t, SuggestionResult{
			Mistakes: MistakeTypo,
			Suggestions: []Suggestion{
				{Value: "range", Score: 1.7545288007923614},
			},
		}, result)

		result = s.Suggest("arang", WithMaxSuggestions(5), WithMaxErrors(2))
		require.Equal(t, SuggestionResult{
			Mistakes: MistakeTypo,
			Suggestions: []Suggestion{
				{Value: "orange", Score: 0.2772588722239781},
				{Value: "range", Score: 0.13862943611198905},
			},
		}, result)
	})

	t.Run("max suggestions", func(t *testing.T) {
		t.Parallel()

		s, err := New(EN)
		require.NoError(t, err)
		s.Add("ac", "ad", "ae", "af", "ag", "ah", "ai", "aj", "ak", "al", "am", "an")

		require.Len(t, s.Suggest("ab").Suggestions, defaultMaxSuggestions)
		require.Len(t, s.Suggest("ab", WithMaxSuggestions(3)).Suggestions, 3)
		require.Len(t, s.Suggest("ab", WithMaxSuggestions(0)).Suggestions, defaultMaxSuggestions)
		require.Len(t, s.Suggest("ab", WithMaxSuggestions(-1)).Suggestions, defaultMaxSuggestions)
	})

	t.Run("valid word", func(t *testing.T) {
		t.Parallel()

		s := newSampleSpellchecker(t)
		result := s.Suggest("orange", WithMaxSuggestions(5))
		require.Equal(t, SuggestionResult{Mistakes: NoMistake}, result)
	})

	t.Run("unknown word", func(t *testing.T) {
		t.Parallel()

		s := newSampleSpellchecker(t)
		result := s.Suggest("qwerty", WithMaxSuggestions(5))
		require.Equal(t, SuggestionResult{Mistakes: MistakeUnknownWord}, result)
	})
}
