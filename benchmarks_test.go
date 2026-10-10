package spellchecker

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Benchmark_Spellchecker_AddPhrases(b *testing.B) {
	for b.Loop() {
		newFullSpellchecker(b)
	}
}

func Benchmark_Spellchecker_IsCorrect(b *testing.B) {
	m := loadFullSpellchecker(b)

	b.ResetTimer()

	for b.Loop() {
		m.IsCorrect("tea")
	}
}

func Benchmark_Spellchecker_Suggest_3(b *testing.B) {
	m := loadFullSpellchecker(b)

	b.ResetTimer()

	for b.Loop() {
		m.Suggest("tee", 5)
	}
}

func Benchmark_Spellchecker_Fix_6_Transposition(b *testing.B) {
	m := loadFullSpellchecker(b)

	b.ResetTimer()

	for b.Loop() {
		m.Suggest("oragne", 5)
	}
}

func Benchmark_Spellchecker_Fix_6_Replacement(b *testing.B) {
	m := loadFullSpellchecker(b)

	b.ResetTimer()

	for b.Loop() {
		m.Suggest("problam", 5)
	}
}

func Benchmark_LayoutCorrector_Correct(b *testing.B) {
	unknown := SuggestionResult{Mistakes: MistakeUnknownWord}
	results := map[string]SuggestionResult{
		"ghbdtn":      unknown,
		"ghbdtnn":     unknown,
		"руддщбцщкдв": unknown,
		"qwzx":        unknown,
		"йцяч":        unknown,
		"12345":       unknown,
		"приветт": {
			Mistakes:    MistakeTypo,
			Suggestions: []Suggestion{{Value: "привет", Score: 1}},
		},
	}

	tests := []struct {
		name   string
		phrase string
	}{
		{name: "correct", phrase: "hello world привет мир"},
		{name: "layout", phrase: "ghbdtn"},
		{name: "layout_split_with_gap", phrase: "руддщбцщкдв"},
		{name: "layout_and_typo", phrase: "ghbdtnn"},
		{name: "switched_unknown", phrase: "qwzx"},
		{name: "unchanged_by_layout", phrase: "12345"},
		{name: "phrase", phrase: "hello ghbdtn мир руддщбцщкдв 12345 ghbdtnn qwzx"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			segments := benchmarkSegments(tt.phrase, results)
			tokenizer := newTokenizerMock(NewStandardTokenizer(), QwertyRuEn, tt.phrase)
			f := NewLayoutCorrector(QwertyRuEn, &spellcheckerMock{results: results}, tokenizer)

			b.ReportAllocs()

			for b.Loop() {
				f.Correct(tt.phrase, segments, 5)
			}
		})
	}
}

// tokenizerMock returns prepared tokens. It copies them, because LayoutCorrector modifies tokens.
type tokenizerMock struct {
	tokens map[string][]Token
	buf    []Token
}

// newTokenizerMock prepares tokens for the words of the phrase in the switched layout.
func newTokenizerMock(tokenizer Tokenizer, replacer replacer, phrase string) *tokenizerMock {
	m := &tokenizerMock{tokens: make(map[string][]Token)}

	for word := range strings.FieldsSeq(phrase) {
		switched := replacer.Replace(word)
		m.tokens[switched] = tokenizer.Tokenize(switched)
	}

	return m
}

func (m *tokenizerMock) Tokenize(input string) []Token {
	m.buf = append(m.buf[:0], m.tokens[input]...)

	return m.buf
}

// benchmarkSegments splits the phrase by spaces into segments.
func benchmarkSegments(phrase string, results map[string]SuggestionResult) []Segment {
	var segments []Segment //nolint:prealloc

	start := 0
	for word := range strings.SplitSeq(phrase, " ") {
		end := start + len(word)
		segments = append(segments, Segment{
			Start:       start,
			End:         end,
			Suggestions: results[word].Suggestions,
			Mistakes:    results[word].Mistakes,
		})
		start = end + 1
	}

	return segments
}

func Benchmark_Norvig1(b *testing.B) {
	benchmarkNorvig(b, "data/norvig1.txt")
}

func Benchmark_Norvig2(b *testing.B) {
	benchmarkNorvig(b, "data/norvig2.txt")
}

type benchmarkNorvigItem struct {
	expected string
	words    []string
}

func benchmarkNorvig(b *testing.B, dataPath string) {
	b.Helper()

	b.StopTimer()
	b.ResetTimer()

	m := loadFullSpellchecker(b)

	testData, err := os.Open(dataPath)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(testData)
	scanner.Split(bufio.ScanLines)

	var data []benchmarkNorvigItem

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			panic(err)
		}

		line := scanner.Text()

		parts := strings.Split(line, ":")
		required := parts[0]
		checks := strings.Split(parts[1], " ")

		data = append(data, benchmarkNorvigItem{
			expected: required,
			words:    checks,
		})
	}

	total := 0
	ok := 0

	for i := range b.N {
		for _, item := range data {
			for _, word := range item.words {
				if word == "" {
					continue
				}

				b.StartTimer()

				result := m.Suggest(word, 10)

				b.StopTimer()

				if i == 0 {
					total++

					if result.Mistakes == NoMistake && word == item.expected {
						ok++
						continue
					}

					if len(result.Suggestions) > 0 && result.Suggestions[0].Value == item.expected {
						ok++
						continue
					}

					// got := ""
					// if len(result) > 0 {
					// 	got = result[0]
					// }

					// b.Logf(
					// 	"word %q: expected %q, got %s, all: %v\n",
					// 	word, item.expected, got, result,
					// )
				}
			}
		}

		b.ReportMetric(float64(ok), "success_words")
		b.ReportMetric(float64(total), "total_words")
		b.ReportMetric(float64(ok)/float64(total)*100, "success_percent")
	}
}

func loadFullSpellchecker(tb testing.TB) *Spellchecker {
	tb.Helper()

	var s *Spellchecker

	ff, err := os.Open("data/spellchecker.bin")
	if err == nil {
		s, err = Load(ff)
		require.NoError(tb, err)

		return s
	}

	s = newFullSpellchecker(tb)

	dst, err := os.Create("data/spellchecker.bin")
	require.NoError(tb, err)

	err = s.Save(dst)
	require.NoError(tb, err)

	return s
}

func newFullSpellchecker(tb testing.TB) *Spellchecker {
	tb.Helper()

	s, err := New(EN)
	require.NoError(tb, err)

	for _, token := range NewWhitespaceTokenizer().Tokenize(strings.ToLower(string(mustReadFile(tb, "data/big.txt")))) {
		s.Add(token.Text)
	}

	return s
}

func newSampleSpellchecker(tb testing.TB) *Spellchecker {
	tb.Helper()

	s, err := New(EN)
	require.NoError(tb, err)

	for _, token := range NewWhitespaceTokenizer().Tokenize(strings.ToLower(string(mustReadFile(tb, "data/sample.txt")))) {
		s.Add(token.Text)
	}

	return s
}

func mustReadFile(tb testing.TB, path string) []byte {
	tb.Helper()

	b, err := os.ReadFile(path)
	require.NoError(tb, err)

	return b
}
