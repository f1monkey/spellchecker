package spellchecker

type suggester interface {
	Suggest(word string, n int, opts ...OptionFunc) SuggestionResult
	AddWeight(weight uint, words ...string)
}

// Mistake is a bit set of mistakes found in a segment or fixed by a suggestion.
// A single segment may contain several mistakes at once,
// e.g. a wrong keyboard layout and a typo.
type Mistake int

// NoMistake means the segment is correct.
const NoMistake Mistake = 0

const (
	// MistakeTypo is a misspelled word (insertion, deletion, substitution or transposition).
	MistakeTypo Mistake = 1 << iota
	// MistakeUnknownWord is a word that is not in the dictionary and has no fix suggestions.
	// It is never set together with other mistakes.
	MistakeUnknownWord
)

// Has reports whether m contains any of the mistakes in x.
// Has(NoMistake) is always false; compare with NoMistake to check for a correct segment.
func (m Mistake) Has(x Mistake) bool { return m&x != 0 }

// Segment is a part of the input phrase: either a correct word or a fragment that needs fixing.
// Start and End are byte offsets in the input phrase, End is exclusive,
// so phrase[Start:End] is the segment text.
// For example, if "ghbdtn" at the beginning of the phrase is replaced with "привет",
// then Start is 0 and End is 6; for "руддщ" replaced with "hello" End is 10.
type Segment struct {
	Start int
	End   int
	// Suggestions are fix candidates ordered from best to worst. Empty for correct segments.
	Suggestions []FixSuggestion
	// Mistakes is a union of Mistakes of all Suggestions, or MistakeUnknownWord if there are none.
	// NoMistake for correct segments.
	Mistakes Mistake
}

// FixSuggestion is a fix candidate for a segment.
type FixSuggestion struct {
	Suggestion
	// Mistakes is a set of mistakes fixed by this suggestion.
	Mistakes Mistake
}

// PhraseFixResult is the result of PhraseFixer.Fix.
type PhraseFixResult struct {
	// Segments cover the whole input phrase in order.
	Segments []Segment
}

// PhraseFixer fixes phrases: in addition to word typos it handles
// missing and extra spaces and wrong keyboard layout.
// It uses Spellchecker to look up and fix individual words.
type PhraseFixer struct {
	spellchecker suggester
	tokenizer    Tokenizer
}

// NewPhraseFixer creates a PhraseFixer that uses the given spellchecker for word lookup.
func NewPhraseFixer(
	spellchecker suggester,
	tokenizer Tokenizer,
) *PhraseFixer {
	return &PhraseFixer{
		spellchecker: spellchecker,
		tokenizer:    tokenizer,
	}
}

// AddPhrases tokenizes each phrase with the spellchecker's tokenizer and adds
// the resulting words with weight 1.
func (f *PhraseFixer) AddPhrases(phrases ...string) {
	f.AddPhraseWeight(1, phrases...)
}

// AddPhraseWeight is like AddPhrases, but each token is added with the given
// weight.
func (f *PhraseFixer) AddPhraseWeight(weight uint, phrases ...string) {
	for _, phrase := range phrases {
		for _, token := range f.tokenizer.Tokenize(phrase) {
			f.spellchecker.AddWeight(weight, token.Text)
		}
	}
}

// Fix splits the phrase into segments and finds mistakes and fix suggestions for each of them.
func (f *PhraseFixer) Fix(phrase string, n int, opts ...OptionFunc) PhraseFixResult {
	tokens := f.tokenizer.Tokenize(phrase)
	if len(tokens) == 0 {
		return PhraseFixResult{}
	}

	segments := make([]Segment, 0, len(tokens))

	for _, token := range tokens {
		suggestions := f.spellchecker.Suggest(token.Text, n, opts...)

		segment := Segment{
			Start:    token.Start,
			End:      token.End,
			Mistakes: NoMistake,
		}

		if !suggestions.ExactMatch {
			if len(suggestions.Suggestions) > 0 {
				fixes := make([]FixSuggestion, 0, len(suggestions.Suggestions))
				for _, s := range suggestions.Suggestions {
					fixes = append(fixes, FixSuggestion{
						Mistakes:   MistakeTypo,
						Suggestion: s,
					})
				}

				segment.Suggestions = fixes
				segment.Mistakes = fixes[0].Mistakes
			} else {
				segment.Mistakes = MistakeUnknownWord
			}
		}

		segments = append(segments, segment)
	}

	return PhraseFixResult{
		Segments: segments,
	}
}
