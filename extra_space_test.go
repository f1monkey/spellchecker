package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewExtraSpaceCorrector(t *testing.T) {
	t.Parallel()

	c := NewExtraSpaceCorrector(3)
	require.Equal(t, 3, c.maxJoin)
}

func Test_ExtraSpaceCorrector_Correct(t *testing.T) {
	t.Parallel()

	typo := SuggestionResult{Mistakes: MistakeTypo, Suggestions: []Suggestion{{Value: "hello", Score: 1}}}
	unknown := SuggestionResult{Mistakes: MistakeUnknownWord}

	tests := []struct {
		name    string
		results map[string]SuggestionResult
		maxJoin int
		phrase  string
		want    []segmentView
	}{
		{
			name:   "empty",
			phrase: "",
			want:   []segmentView{},
		},
		{
			name:   "correct words are not joined",
			phrase: "in to",
			want: []segmentView{
				{Text: "in", Mistakes: NoMistake},
				{Text: "to", Mistakes: NoMistake},
			},
		},
		{
			name:    "words with mistakes are joined",
			results: map[string]SuggestionResult{"hel": typo, "lo": unknown},
			phrase:  "hel lo",
			want: []segmentView{
				{Text: "hel lo", Suggestions: []string{"hello"}, Mistakes: MistakeExtraSpace},
			},
		},
		{
			name:    "correct word is joined with a wrong one",
			results: map[string]SuggestionResult{"pple": unknown},
			phrase:  "a pple",
			want: []segmentView{
				{Text: "a pple", Suggestions: []string{"apple"}, Mistakes: MistakeExtraSpace},
			},
		},
		{
			name:    "wrong word is joined with a previous correct one",
			results: map[string]SuggestionResult{"pple": unknown, "theapple": unknown},
			phrase:  "the a pple",
			want: []segmentView{
				{Text: "the", Mistakes: NoMistake},
				{Text: "a pple", Suggestions: []string{"apple"}, Mistakes: MistakeExtraSpace},
			},
		},
		{
			name:    "three segments are joined",
			results: map[string]SuggestionResult{"l": unknown, "lo": unknown},
			phrase:  "he l lo",
			want: []segmentView{
				{Text: "he l lo", Suggestions: []string{"hello"}, Mistakes: MistakeExtraSpace},
			},
		},
		{
			name:    "longest join wins",
			results: map[string]SuggestionResult{"wi": unknown, "th": unknown},
			phrase:  "wi th out",
			want: []segmentView{
				{Text: "wi th out", Suggestions: []string{"without"}, Mistakes: MistakeExtraSpace},
			},
		},
		{
			name:    "join is limited by maxJoin",
			results: map[string]SuggestionResult{"wi": unknown, "th": unknown},
			maxJoin: 2,
			phrase:  "wi th out",
			want: []segmentView{
				{Text: "wi th", Suggestions: []string{"with"}, Mistakes: MistakeExtraSpace},
				{Text: "out", Mistakes: NoMistake},
			},
		},
		{
			name:    "several spaces and tabs",
			results: map[string]SuggestionResult{"hel": typo},
			phrase:  "hel \t lo",
			want: []segmentView{
				{Text: "hel \t lo", Suggestions: []string{"hello"}, Mistakes: MistakeExtraSpace},
			},
		},
		{
			name:    "joined through punctuation",
			results: map[string]SuggestionResult{"hel": typo},
			phrase:  "hel,lo",
			want: []segmentView{
				{Text: "hel,lo", Suggestions: []string{"hello"}, Mistakes: MistakeExtraSpace},
			},
		},
		{
			name:    "joined through punctuation and spaces",
			results: map[string]SuggestionResult{"hel": typo},
			phrase:  "hel. lo",
			want: []segmentView{
				{Text: "hel. lo", Suggestions: []string{"hello"}, Mistakes: MistakeExtraSpace},
			},
		},
		{
			name:   "correct words are not joined through punctuation",
			phrase: "in,to",
			want: []segmentView{
				{Text: "in", Mistakes: NoMistake},
				{Text: "to", Mistakes: NoMistake},
			},
		},
		{
			name:    "unknown joined word is not used",
			results: map[string]SuggestionResult{"hel": typo, "hello": unknown},
			phrase:  "hel lo",
			want: []segmentView{
				{Text: "hel", Suggestions: []string{"hello"}, Mistakes: MistakeTypo},
				{Text: "lo", Mistakes: NoMistake},
			},
		},
		{
			name:    "joined typo is not used",
			results: map[string]SuggestionResult{"hel": typo, "hello": typo},
			phrase:  "hel lo",
			want: []segmentView{
				{Text: "hel", Suggestions: []string{"hello"}, Mistakes: MistakeTypo},
				{Text: "lo", Mistakes: NoMistake},
			},
		},
		{
			name: "order and offsets in a phrase",
			results: map[string]SuggestionResult{
				"hel": typo, "lo": unknown, "wor": unknown, "ld": unknown,
				"foohel": unknown, "foohello": unknown, "hellowor": unknown, "loworld": unknown,
			},
			phrase: "foo hel lo wor ld",
			want: []segmentView{
				{Text: "foo", Mistakes: NoMistake},
				{Text: "hel lo", Suggestions: []string{"hello"}, Mistakes: MistakeExtraSpace},
				{Text: "wor ld", Suggestions: []string{"world"}, Mistakes: MistakeExtraSpace},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			maxJoin := tt.maxJoin
			if maxJoin == 0 {
				maxJoin = 3
			}

			sc := &spellcheckerMock{results: tt.results}
			f := NewPhraseFixer(sc, NewStandardTokenizer(), NewExtraSpaceCorrector(maxJoin))

			require.Equal(t, tt.want, segmentViews(t, tt.phrase, f.Fix(tt.phrase, WithMaxSuggestions(3))))
		})
	}

	t.Run("segment offsets are in bytes of the phrase", func(t *testing.T) {
		t.Parallel()

		c := NewExtraSpaceCorrector(3)
		segments := []Segment{
			{Text: "ок", Start: 0, End: 4, Mistakes: NoMistake},
			{Text: "при", Start: 5, End: 11, Mistakes: MistakeUnknownWord},
			{Text: "вет", Start: 12, End: 18, Mistakes: MistakeUnknownWord},
		}

		require.Equal(t, []Segment{
			{Text: "ок при вет", Start: 0, End: 18, Suggestions: []Suggestion{{Value: "окпривет"}}, Mistakes: MistakeExtraSpace},
		}, c.Correct(CorrectInput{Phrase: "ок при вет", Tokenizer: NewStandardTokenizer(), Spellchecker: &spellcheckerMock{}}, segments))
	})
}

func Test_ExtraSpaceCorrector_Correct_Integration(t *testing.T) {
	t.Parallel()

	s, err := New(EN, RU)
	require.NoError(t, err)
	s.Add("hello", "world", "in", "to", "into", "привет")

	phrase := "hel lo in to wor ld ghbdtn"
	f := NewPhraseFixer(s, NewStandardTokenizer(),
		NewExtraSpaceCorrector(3),
		NewLayoutCorrector(QwertyRuEn()),
	)

	require.Equal(t, []segmentView{
		{Text: "hel lo", Suggestions: []string{"hello"}, Mistakes: MistakeExtraSpace},
		{Text: "in", Mistakes: NoMistake},
		{Text: "to", Mistakes: NoMistake},
		{Text: "wor ld", Suggestions: []string{"world"}, Mistakes: MistakeExtraSpace},
		{Text: "ghbdtn", Suggestions: []string{"привет"}, Mistakes: MistakeLayout},
	}, segmentViews(t, phrase, f.Fix(phrase, WithMaxSuggestions(3))))
}
