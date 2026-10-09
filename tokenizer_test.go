package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWhitespaceTokenizer(t *testing.T) {
	t.Parallel()

	tok := NewWhitespaceTokenizer()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty",
			input: "",
			want:  []string{},
		},
		{
			name:  "single word",
			input: "hello",
			want:  []string{"hello"},
		},
		{
			name:  "spaces",
			input: "hello world",
			want:  []string{"hello", "world"},
		},
		{
			name:  "leading trailing and repeated whitespace",
			input: "  hello   \tworld \n ",
			want:  []string{"hello", "world"},
		},
		{
			name:  "keeps punctuation and hyphens",
			input: "The 2 QUICK Brown-Foxes jumped over the lazy dog's bone.",
			want:  []string{"The", "2", "QUICK", "Brown-Foxes", "jumped", "over", "the", "lazy", "dog's", "bone."},
		},
		{
			name:  "unicode",
			input: "привет  мир",
			want:  []string{"привет", "мир"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tokenTexts(t, tt.input, tok.Tokenize(tt.input)))
		})
	}
}

func TestStandardTokenizer(t *testing.T) {
	t.Parallel()

	tok := NewStandardTokenizer()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty",
			input: "",
			want:  []string{},
		},
		{
			name:  "elasticsearch example",
			input: "The 2 QUICK Brown-Foxes jumped over the lazy dog's bone.",
			want:  []string{"The", "2", "QUICK", "Brown", "Foxes", "jumped", "over", "the", "lazy", "dog's", "bone"},
		},
		{
			name:  "apostrophe and curly apostrophe",
			input: "You're dog’s",
			want:  []string{"You're", "dog’s"},
		},
		{
			name:  "underscores stay",
			input: "jumps_over the fence",
			want:  []string{"jumps_over", "the", "fence"},
		},
		{
			name:  "email splits like standard not uax_url_email",
			input: "write to somebody@example.com please",
			want:  []string{"write", "to", "somebody", "example", "com", "please"},
		},
		{
			name:  "unicode letters",
			input: "Мама мыла раму, папа-кот.",
			want:  []string{"Мама", "мыла", "раму", "папа", "кот"},
		},
		{
			name:  "alphanumeric",
			input: "I am the 1st",
			want:  []string{"I", "am", "the", "1st"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tokenTexts(t, tt.input, tok.Tokenize(tt.input)))
		})
	}
}

func TestTokenizerOffsets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tok   Tokenizer
		input string
		want  []Token
	}{
		{
			name:  "whitespace ascii",
			tok:   NewWhitespaceTokenizer(),
			input: "  hello \tworld ",
			want: []Token{
				{Text: "hello", Start: 2, End: 7},
				{Text: "world", Start: 9, End: 14},
			},
		},
		{
			name:  "whitespace multibyte",
			tok:   NewWhitespaceTokenizer(),
			input: "привет 😀 мир",
			want: []Token{
				{Text: "привет", Start: 0, End: 12},
				{Text: "😀", Start: 13, End: 17},
				{Text: "мир", Start: 18, End: 24},
			},
		},
		{
			name:  "standard punctuation and emoji",
			tok:   NewStandardTokenizer(),
			input: "папа-кот😀dog’s",
			want: []Token{
				{Text: "папа", Start: 0, End: 8},
				{Text: "кот", Start: 9, End: 15},
				{Text: "dog’s", Start: 19, End: 26},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tt.tok.Tokenize(tt.input))
		})
	}
}

func tokenTexts(t *testing.T, input string, tokens []Token) []string {
	t.Helper()

	texts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		require.Equal(t, token.Text, input[token.Start:token.End])
		texts = append(texts, token.Text)
	}

	return texts
}
