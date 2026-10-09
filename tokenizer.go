package spellchecker

import "regexp"

// Token is a fragment of the tokenized input.
// Start and End are byte offsets in the input, End is exclusive,
// so input[Start:End] == Text.
type Token struct {
	Text  string
	Start int
	End   int
}

type Tokenizer interface {
	Tokenize(input string) []Token
}

// NewWhitespaceTokenizer splits on Unicode whitespace, like Elasticsearch whitespace tokenizer.
func NewWhitespaceTokenizer() *RegexpTokenizer {
	return NewRegexpTokenizer(regexp.MustCompile(`\S+`))
}

// NewStandardTokenizer approximates Elasticsearch standard tokenizer (UAX #29):
// keeps letters, digits, underscores and in-word apostrophes; splits on hyphens
// and other punctuation.
func NewStandardTokenizer() *RegexpTokenizer {
	return NewRegexpTokenizer(regexp.MustCompile(`[\p{L}\p{N}_'’]+`))
}

type RegexpTokenizer struct {
	regexp *regexp.Regexp
}

// NewRegexpTokenizer returns a tokenizer that emits every non-empty match of regexp as a token.
// Note that regexp matches tokens, not separators.
func NewRegexpTokenizer(regexp *regexp.Regexp) *RegexpTokenizer {
	return &RegexpTokenizer{
		regexp: regexp,
	}
}

func (t *RegexpTokenizer) Tokenize(input string) []Token {
	matches := t.regexp.FindAllStringIndex(input, -1)
	tokens := make([]Token, 0, len(matches))

	for _, m := range matches {
		start, end := m[0], m[1]
		if start == end {
			continue
		}

		tokens = append(tokens, Token{
			Text:  input[start:end],
			Start: start,
			End:   end,
		})
	}

	return tokens
}
