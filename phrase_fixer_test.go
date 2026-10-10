package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewPhraseFixer(t *testing.T) {
	t.Parallel()

	sc := &spellcheckerMock{}
	tok := NewWhitespaceTokenizer()
	lf := NewLayoutFixer(QwertyRuEn, sc, tok)

	f := NewPhraseFixer(sc, tok, lf)
	require.Same(t, sc, f.spellchecker)
	require.Same(t, tok, f.tokenizer)
	require.Same(t, lf, f.fixer)
}

func Test_PhraseFixer_Fix_Fixer(t *testing.T) {
	t.Parallel()

	sc := &spellcheckerMock{results: map[string]SuggestionResult{"helo": {Mistakes: MistakeTypo}}}
	fixer := &segmentFixerMock{results: map[string][]Segment{
		"hello world": {{Start: 0, End: 11, Mistakes: MistakeLayout}},
	}}
	f := NewPhraseFixer(sc, NewWhitespaceTokenizer(), fixer)

	tests := []struct {
		name   string
		phrase string
		want   []Segment
	}{
		{
			name:   "segments are passed to the fixer",
			phrase: "helo world",
			want: []Segment{
				{Start: 0, End: 4, Mistakes: MistakeTypo},
				{Start: 5, End: 10, Mistakes: NoMistake},
			},
		},
		{
			name:   "fixer result is returned",
			phrase: "hello world",
			want:   []Segment{{Start: 0, End: 11, Mistakes: MistakeLayout}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, PhraseFixResult{Segments: tt.want}, f.Fix(tt.phrase, 3))
		})
	}
}

func Test_PhraseFixer_AddPhrases(t *testing.T) {
	t.Parallel()

	sc := &spellcheckerMock{}
	f := NewPhraseFixer(sc, NewStandardTokenizer(), nil)

	f.AddPhrases("hello, world", "", "папа-кот", "hello")

	require.Equal(t, map[string]uint{
		"hello": 2,
		"world": 1,
		"папа":  1,
		"кот":   1,
	}, sc.weights)
}

func Test_PhraseFixer_AddPhraseWeight(t *testing.T) {
	t.Parallel()

	sc := &spellcheckerMock{}
	f := NewPhraseFixer(sc, NewWhitespaceTokenizer(), nil)

	f.AddPhraseWeight(5, "foo bar")

	require.Equal(t, map[string]uint{
		"foo": 5,
		"bar": 5,
	}, sc.weights)
}

func Test_PhraseFixer_Fix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		results map[string]SuggestionResult
		phrase  string
		want    PhraseFixResult
	}{
		{
			name:   "empty phrase",
			phrase: "",
			want:   PhraseFixResult{},
		},
		{
			name: "all words are correct",
			results: map[string]SuggestionResult{
				"hello": {},
				"world": {},
			},
			phrase: "hello world",
			want: PhraseFixResult{Segments: []Segment{
				{Start: 0, End: 5, Mistakes: NoMistake},
				{Start: 6, End: 11, Mistakes: NoMistake},
			}},
		},
		{
			name: "typo with suggestions",
			results: map[string]SuggestionResult{
				"helo": {
					Mistakes:    MistakeTypo,
					Suggestions: []Suggestion{{Value: "hello", Score: 2}, {Value: "help", Score: 1}},
				},
				"world": {},
			},
			phrase: "helo world",
			want: PhraseFixResult{Segments: []Segment{
				{
					Start:       0,
					End:         4,
					Suggestions: []Suggestion{{Value: "hello", Score: 2}, {Value: "help", Score: 1}},
					Mistakes:    MistakeTypo,
				},
				{Start: 5, End: 10, Mistakes: NoMistake},
			}},
		},
		{
			name: "unknown word without suggestions",
			results: map[string]SuggestionResult{
				"qwzx": {Mistakes: MistakeUnknownWord},
			},
			phrase: "qwzx",
			want: PhraseFixResult{Segments: []Segment{
				{Start: 0, End: 4, Mistakes: MistakeUnknownWord},
			}},
		},
		{
			name: "byte offsets for multibyte text and extra whitespace",
			results: map[string]SuggestionResult{
				"привет": {},
				"мир":    {},
			},
			phrase: "  привет \t мир ",
			want: PhraseFixResult{Segments: []Segment{
				{Start: 2, End: 14, Mistakes: NoMistake},
				{Start: 17, End: 23, Mistakes: NoMistake},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := NewPhraseFixer(&spellcheckerMock{results: tt.results}, NewWhitespaceTokenizer(), nil)

			got := f.Fix(tt.phrase, 3)
			require.Equal(t, tt.want, got)

			for _, s := range got.Segments {
				require.LessOrEqual(t, s.Start, s.End)
				require.LessOrEqual(t, s.End, len(tt.phrase))
			}
		})
	}
}

func Test_PhraseFixer_Fix_Integration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tokenizer Tokenizer
		phrase    string
		n         int
		opts      []OptionFunc
		want      []segmentView
	}{
		{
			name:      "empty phrase",
			tokenizer: NewWhitespaceTokenizer(),
			phrase:    "",
			n:         5,
			want:      []segmentView{},
		},
		{
			name:      "correct phrase",
			tokenizer: NewWhitespaceTokenizer(),
			phrase:    "green tea",
			n:         5,
			want: []segmentView{
				{Text: "green", Mistakes: NoMistake},
				{Text: "tea", Mistakes: NoMistake},
			},
		},
		{
			name:      "typo",
			tokenizer: NewWhitespaceTokenizer(),
			phrase:    "arang",
			n:         5,
			want: []segmentView{
				{Text: "arang", Suggestions: []string{"orange", "range"}, Mistakes: MistakeTypo},
			},
		},
		{
			name:      "n limits suggestions",
			tokenizer: NewWhitespaceTokenizer(),
			phrase:    "arang",
			n:         1,
			want: []segmentView{
				{Text: "arang", Suggestions: []string{"orange"}, Mistakes: MistakeTypo},
			},
		},
		{
			name:      "options are applied",
			tokenizer: NewWhitespaceTokenizer(),
			phrase:    "rang",
			n:         5,
			opts:      []OptionFunc{WithMaxErrors(1)},
			want: []segmentView{
				{Text: "rang", Suggestions: []string{"range"}, Mistakes: MistakeTypo},
			},
		},
		{
			name:      "unknown word",
			tokenizer: NewWhitespaceTokenizer(),
			phrase:    "qwerty",
			n:         5,
			want: []segmentView{
				{Text: "qwerty", Mistakes: MistakeUnknownWord},
			},
		},
		{
			name:      "correct, typo and unknown words together",
			tokenizer: NewWhitespaceTokenizer(),
			phrase:    "  blak   tea\tqwerty ",
			n:         5,
			want: []segmentView{
				{Text: "blak", Suggestions: []string{"black"}, Mistakes: MistakeTypo},
				{Text: "tea", Mistakes: NoMistake},
				{Text: "qwerty", Mistakes: MistakeUnknownWord},
			},
		},
		{
			name:      "standard tokenizer skips punctuation",
			tokenizer: NewStandardTokenizer(),
			phrase:    "green-tea, cofee!",
			n:         5,
			want: []segmentView{
				{Text: "green", Mistakes: NoMistake},
				{Text: "tea", Mistakes: NoMistake},
				{Text: "cofee", Suggestions: []string{"coffee"}, Mistakes: MistakeTypo},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := NewPhraseFixer(newSampleSpellchecker(t), tt.tokenizer, nil)

			got := f.Fix(tt.phrase, tt.n, tt.opts...)
			require.Equal(t, tt.want, segmentViews(t, tt.phrase, got))
		})
	}
}

type spellcheckerMock struct {
	results map[string]SuggestionResult
	weights map[string]uint
}

func (m *spellcheckerMock) Suggest(word string, _ int, _ ...OptionFunc) SuggestionResult {
	return m.results[word]
}

func (m *spellcheckerMock) AddWeight(weight uint, words ...string) {
	if m.weights == nil {
		m.weights = make(map[string]uint)
	}

	for _, w := range words {
		m.weights[w] += weight
	}
}

// segmentFixerMock returns segments as is for unknown phrases.
type segmentFixerMock struct {
	results map[string][]Segment
}

func (m *segmentFixerMock) Fix(phrase string, segments []Segment, _ int, _ ...OptionFunc) []Segment {
	if result, ok := m.results[phrase]; ok {
		return result
	}

	return segments
}

// segmentView is a Segment with its text and suggestion values instead of offsets and scores.
type segmentView struct {
	Text        string
	Suggestions []string
	Mistakes    Mistake
}

func segmentViews(t *testing.T, phrase string, result PhraseFixResult) []segmentView {
	t.Helper()

	views := make([]segmentView, 0, len(result.Segments))
	for _, s := range result.Segments {
		v := segmentView{
			Text:     phrase[s.Start:s.End],
			Mistakes: s.Mistakes,
		}
		for _, suggestion := range s.Suggestions {
			v.Suggestions = append(v.Suggestions, suggestion.Value)
		}

		views = append(views, v)
	}

	return views
}
