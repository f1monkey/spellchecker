package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Spellchecker_SuggestScore(t *testing.T) {
	t.Run("fix", func(t *testing.T) {
		s := newSampleSpellchecker()
		result := s.Suggest("arang", 5)
		require.Equal(t, SuggestionResult{
			Suggestions: []Match{
				{Value: "orange", Score: 0.2772588722239781},
				{Value: "range", Score: 0.13862943611198905},
			},
		}, result)
	})

	t.Run("custom max errors", func(t *testing.T) {
		s := newSampleSpellchecker()
		result := s.Suggest("arang", 5, SuggestWithMaxErrors(1))
		require.Equal(t, SuggestionResult{
			Suggestions: []Match{
				{Value: "range", Score: 0.13862943611198905},
			},
		}, result)

		result = s.Suggest("arang", 5, SuggestWithMaxErrors(2))
		require.Equal(t, SuggestionResult{
			Suggestions: []Match{
				{Value: "orange", Score: 0.2772588722239781},
				{Value: "range", Score: 0.13862943611198905},
			},
		}, result)
	})

	t.Run("valid word", func(t *testing.T) {
		s := newSampleSpellchecker()
		result := s.Suggest("orange", 5)
		require.Equal(t, SuggestionResult{ExactMatch: true}, result)
	})

	t.Run("unknown word", func(t *testing.T) {
		s := newSampleSpellchecker()
		result := s.Suggest("qwerty", 5)
		require.Equal(t, SuggestionResult{Suggestions: []Match{}}, result)
	})
}
