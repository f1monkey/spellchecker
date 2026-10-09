package spellchecker

type suggester interface {
	Suggest(word string, n int, opts ...OptionFunc) SuggestionResult
	AddWeight(weight uint, words ...string)
}

// Segment is a part of the input phrase: either a correct word or a fragment that needs fixing.
// Start and End are byte offsets in the input phrase, End is exclusive,
// so phrase[Start:End] is the segment text.
// For example, if "ghbdtn" at the beginning of the phrase is replaced with "привет",
// then Start is 0 and End is 6; for "руддщ" replaced with "hello" End is 10.
type Segment struct {
	Start int
	End   int
	// Suggestions are fix candidates ordered from best to worst. Empty for correct and unknown segments.
	Suggestions []Suggestion
	// Mistakes is a set of mistakes fixed by Suggestions, or MistakeUnknownWord if there are none.
	// NoMistake for correct segments.
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
		suggestionResult := f.spellchecker.Suggest(token.Text, n, opts...)

		segment := Segment{
			Start:       token.Start,
			End:         token.End,
			Suggestions: suggestionResult.Suggestions,
			Mistakes:    suggestionResult.Mistakes,
		}

		segments = append(segments, segment)
	}

	return PhraseFixResult{
		Segments: segments,
	}
}
