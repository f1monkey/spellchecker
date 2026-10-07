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

			require.Equal(t, tt.want, tok.Tokenize(tt.input))
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

			require.Equal(t, tt.want, tok.Tokenize(tt.input))
		})
	}
}
