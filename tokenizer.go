package spellchecker

import "regexp"

type Tokenizer interface {
	Tokenize(input string) []string
}

// NewWhitespaceTokenizer splits on Unicode whitespace, like Elasticsearch whitespace tokenizer.
func NewWhitespaceTokenizer() *RegexpTokenizer {
	return NewRegexpTokenizer(regexp.MustCompile(`\s+`))
}

// NewStandardTokenizer approximates Elasticsearch standard tokenizer (UAX #29):
// keeps letters, digits, underscores and in-word apostrophes; splits on hyphens
// and other punctuation.
func NewStandardTokenizer() *RegexpTokenizer {
	return NewRegexpTokenizer(regexp.MustCompile(`[^\p{L}\p{N}_'’]+`))
}

type RegexpTokenizer struct {
	regexp *regexp.Regexp
}

// NewRegexpTokenizer splits input by regexp and drops empty fragments.
func NewRegexpTokenizer(regexp *regexp.Regexp) *RegexpTokenizer {
	return &RegexpTokenizer{
		regexp: regexp,
	}
}

func (t *RegexpTokenizer) Tokenize(input string) []string {
	parts := t.regexp.Split(input, -1)
	tokens := make([]string, 0, len(parts))

	for _, p := range parts {
		if p == "" {
			continue
		}

		tokens = append(tokens, p)
	}

	return tokens
}
