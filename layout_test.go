package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_QwertyRuEn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "en to ru",
			input: "ghbdtn",
			want:  "привет",
		},
		{
			name:  "ru to en",
			input: "руддщ",
			want:  "hello",
		},
		{
			name:  "en punctuation keys become ru letters",
			input: "j,]tv `krf ;tkt'",
			want:  "объем ёлка желеэ",
		},
		{
			name:  "uppercase",
			input: "Ghbdtn Vbh {JHJIJ",
			want:  "Привет Мир ХОРОШО",
		},
		{
			name:  "shifted punctuation keys",
			input: "~:\"<>{}",
			want:  "ЁЖЭБЮХЪ",
		},
		{
			name:  "mixed layouts in one string",
			input: "ghbdtn руддщ",
			want:  "привет hello",
		},
		{
			name:  "digits, spaces and other symbols stay",
			input: "123 -=!/\t",
			want:  "123 -=!/\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, QwertyRuEn().Replace(tt.input))
		})
	}

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		for _, s := range []string{"ghbdtn", "Объединение", "{jhjij", "`~;:'\",.<>[]{}"} {
			require.Equal(t, s, QwertyRuEn().Replace(QwertyRuEn().Replace(s)))
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		t.Parallel()

		r := QwertyRuEn()
		r['q'] = 'q'

		require.Equal(t, "й", QwertyRuEn().Replace("q"))
	})

	// t.Run("replaces each rune with one rune", func(t *testing.T) {
	// 	t.Parallel()

	// 	_, ok := sourceOffsets("j,]tv `krf", QwertyRuEn().Replace("j,]tv `krf"))
	// 	require.True(t, ok)
	// })
}

func Test_LayoutCorrector_Correct(t *testing.T) {
	t.Parallel()

	typo := func(values ...string) SuggestionResult {
		suggestions := make([]Suggestion, 0, len(values))
		for _, v := range values {
			suggestions = append(suggestions, Suggestion{Value: v, Score: 1})
		}

		return SuggestionResult{Mistakes: MistakeTypo, Suggestions: suggestions}
	}
	unknown := SuggestionResult{Mistakes: MistakeUnknownWord}

	tests := []struct {
		name    string
		results map[string]SuggestionResult
		phrase  string
		want    []segmentView
	}{
		{
			name:   "empty",
			phrase: "",
			want:   []segmentView{},
		},
		{
			name:    "correct words are kept",
			results: map[string]SuggestionResult{"руддщ": unknown},
			phrase:  "hello",
			want: []segmentView{
				{Text: "hello", Mistakes: NoMistake},
			},
		},
		{
			name:    "unknown en to ru",
			results: map[string]SuggestionResult{"ghbdtn": unknown},
			phrase:  "ghbdtn",
			want: []segmentView{
				{Text: "ghbdtn", Suggestions: []string{"привет"}, Mistakes: MistakeLayout},
			},
		},
		{
			name:    "unknown ru to en",
			results: map[string]SuggestionResult{"руддщ": unknown},
			phrase:  "руддщ",
			want: []segmentView{
				{Text: "руддщ", Suggestions: []string{"hello"}, Mistakes: MistakeLayout},
			},
		},
		{
			name:    "switched text is split into tokens",
			results: map[string]SuggestionResult{"руддщбцщкдв": unknown},
			phrase:  "руддщбцщкдв",
			want: []segmentView{
				{Text: "руддщ", Suggestions: []string{"hello"}, Mistakes: MistakeLayout},
				{Text: "б", Suggestions: []string{","}, Mistakes: MistakeLayout},
				{Text: "цщкдв", Suggestions: []string{"world"}, Mistakes: MistakeLayout},
			},
		},
		{
			name:    "text before and after tokens",
			results: map[string]SuggestionResult{"бруддщю": unknown},
			phrase:  "бруддщю",
			want: []segmentView{
				{Text: "б", Suggestions: []string{","}, Mistakes: MistakeLayout},
				{Text: "руддщ", Suggestions: []string{"hello"}, Mistakes: MistakeLayout},
				{Text: "ю", Suggestions: []string{"."}, Mistakes: MistakeLayout},
			},
		},
		{
			name: "text between tokens that does not change has no segment",
			results: map[string]SuggestionResult{
				"руддщ": unknown,
				"цщкдв": unknown,
			},
			phrase: "руддщ-цщкдв",
			want: []segmentView{
				{Text: "руддщ", Suggestions: []string{"hello"}, Mistakes: MistakeLayout},
				{Text: "цщкдв", Suggestions: []string{"world"}, Mistakes: MistakeLayout},
			},
		},
		{
			name: "keys that are punctuation in the original layout",
			results: map[string]SuggestionResult{
				"j":  unknown,
				"tv": unknown,
			},
			phrase: "j,]tv",
			want: []segmentView{
				{Text: "j,]tv", Suggestions: []string{"объем"}, Mistakes: MistakeLayout},
			},
		},
		{
			name: "chunk is kept if switched text is not better",
			results: map[string]SuggestionResult{
				"j":     unknown,
				"tv":    unknown,
				"объем": unknown,
			},
			phrase: "j,]tv",
			want: []segmentView{
				{Text: "j", Mistakes: MistakeUnknownWord},
				{Text: "tv", Mistakes: MistakeUnknownWord},
			},
		},
		{
			name: "layout and typo",
			results: map[string]SuggestionResult{
				"ghbdtnn": unknown,
				"приветт": typo("привет"),
			},
			phrase: "ghbdtnn",
			want: []segmentView{
				{Text: "ghbdtnn", Suggestions: []string{"привет"}, Mistakes: MistakeLayout | MistakeTypo},
			},
		},
		{
			name:    "typo is replaced with correct switched text",
			results: map[string]SuggestionResult{"vbh": typo("vbr")},
			phrase:  "vbh",
			want: []segmentView{
				{Text: "vbh", Suggestions: []string{"мир"}, Mistakes: MistakeLayout},
			},
		},
		{
			name: "typo is kept if switched text is a typo too",
			results: map[string]SuggestionResult{
				"vbh": typo("vbr"),
				"мир": typo("мор"),
			},
			phrase: "vbh",
			want: []segmentView{
				{Text: "vbh", Suggestions: []string{"vbr"}, Mistakes: MistakeTypo},
			},
		},
		{
			name: "unknown is kept if a switched token is unknown",
			results: map[string]SuggestionResult{
				"руддщбцщкдв": unknown,
				"world":       unknown,
			},
			phrase: "руддщбцщкдв",
			want: []segmentView{
				{Text: "руддщбцщкдв", Mistakes: MistakeUnknownWord},
			},
		},
		{
			name:    "unknown is kept if layout does not change it",
			results: map[string]SuggestionResult{"12345": unknown},
			phrase:  "12345",
			want: []segmentView{
				{Text: "12345", Mistakes: MistakeUnknownWord},
			},
		},
		{
			name:    "unknown is kept if switched text has no tokens",
			results: map[string]SuggestionResult{"б": unknown},
			phrase:  "б",
			want: []segmentView{
				{Text: "б", Mistakes: MistakeUnknownWord},
			},
		},
		{
			name: "order and offsets in a phrase",
			results: map[string]SuggestionResult{
				"руддщбцщкдв": unknown,
				"12345":       unknown,
				"ghbdtn":      unknown,
			},
			phrase: " hello  руддщбцщкдв 12345\tghbdtn ",
			want: []segmentView{
				{Text: "hello", Mistakes: NoMistake},
				{Text: "руддщ", Suggestions: []string{"hello"}, Mistakes: MistakeLayout},
				{Text: "б", Suggestions: []string{","}, Mistakes: MistakeLayout},
				{Text: "цщкдв", Suggestions: []string{"world"}, Mistakes: MistakeLayout},
				{Text: "12345", Mistakes: MistakeUnknownWord},
				{Text: "ghbdtn", Suggestions: []string{"привет"}, Mistakes: MistakeLayout},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sc := &spellcheckerMock{results: tt.results}
			f := NewPhraseFixer(sc, NewStandardTokenizer(), NewLayoutCorrector(QwertyRuEn(), sc))

			require.Equal(t, tt.want, segmentViews(t, tt.phrase, f.Fix(tt.phrase, WithMaxSuggestions(3))))
		})
	}

	t.Run("segment offsets are in bytes of the phrase", func(t *testing.T) {
		t.Parallel()

		f := NewLayoutCorrector(QwertyRuEn(), &spellcheckerMock{})

		got := f.correctChunk(Phrase{Text: "ok руддщбцщкдв", Tokenizer: NewStandardTokenizer()}, 3, 25, MistakeUnknownWord)
		require.Equal(t, []Segment{
			{Text: "руддщ", Start: 3, End: 13, Suggestions: []Suggestion{{Value: "hello"}}, Mistakes: MistakeLayout},
			{Text: "б", Start: 13, End: 15, Suggestions: []Suggestion{{Value: ","}}, Mistakes: MistakeLayout},
			{Text: "цщкдв", Start: 15, End: 25, Suggestions: []Suggestion{{Value: "world"}}, Mistakes: MistakeLayout},
		}, got)
	})

	t.Run("segments crossing whitespace are kept", func(t *testing.T) {
		t.Parallel()

		f := NewLayoutCorrector(QwertyRuEn(), &spellcheckerMock{})
		segments := []Segment{
			{Text: "ghbdtn vbh", Start: 0, End: 10, Mistakes: MistakeUnknownWord},
			{Text: "ghbdtn", Start: 11, End: 17, Mistakes: MistakeUnknownWord},
		}

		require.Equal(t, []Segment{
			{Text: "ghbdtn vbh", Start: 0, End: 10, Mistakes: MistakeUnknownWord},
			{Text: "ghbdtn", Start: 11, End: 17, Suggestions: []Suggestion{{Value: "привет"}}, Mistakes: MistakeLayout},
		}, f.Correct(Phrase{Text: "ghbdtn vbh ghbdtn", Tokenizer: NewStandardTokenizer()}, segments))
	})
}

func Test_LayoutCorrector_Correct_Integration(t *testing.T) {
	t.Parallel()

	s, err := New(EN, RU)
	require.NoError(t, err)
	s.Add("hello", "world", "привет", "мир", "объем")

	phrase := "ghbdtn мир руддщбцщкдв ghbdtnn qwzx j,]tv"
	f := NewPhraseFixer(s, NewStandardTokenizer(), NewLayoutCorrector(QwertyRuEn(), s))

	require.Equal(t, []segmentView{
		{Text: "ghbdtn", Suggestions: []string{"привет"}, Mistakes: MistakeLayout},
		{Text: "мир", Mistakes: NoMistake},
		{Text: "руддщ", Suggestions: []string{"hello"}, Mistakes: MistakeLayout},
		{Text: "б", Suggestions: []string{","}, Mistakes: MistakeLayout},
		{Text: "цщкдв", Suggestions: []string{"world"}, Mistakes: MistakeLayout},
		{Text: "ghbdtnn", Suggestions: []string{"привет"}, Mistakes: MistakeLayout | MistakeTypo},
		{Text: "qwzx", Mistakes: MistakeUnknownWord},
		{Text: "j,]tv", Suggestions: []string{"объем"}, Mistakes: MistakeLayout},
	}, segmentViews(t, phrase, f.Fix(phrase, WithMaxSuggestions(3))))
}
