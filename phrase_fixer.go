package spellchecker

type suggester interface {
	Suggest(word string, n int, opts ...OptionFunc) SuggestionResult
	AddWeight(weight uint, words ...string)
}

// Segment is a part of the phrase: a correct word or a fragment to fix.
// Start and End are byte offsets in the phrase.
type Segment struct {
	Start int
	End   int
	// Suggestions are fix candidates, best first.
	Suggestions []Suggestion
	// Mistakes are mistakes found in the segment.
	Mistakes Mistake
}

// PhraseFixResult is the result of PhraseFixer.Fix.
type PhraseFixResult struct {
	// Segments are in phrase order.
	Segments []Segment
}

// SegmentFixer fixes segments of the phrase found by PhraseFixer.
type SegmentFixer interface {
	Fix(phrase string, segments []Segment, n int, opts ...OptionFunc) []Segment
}

// PhraseFixer fixes typos and wrong keyboard layout in phrases.
type PhraseFixer struct {
	spellchecker suggester
	tokenizer    Tokenizer
	fixers       []SegmentFixer
}

// NewPhraseFixer creates a PhraseFixer. Fixers are applied in the given order.
func NewPhraseFixer(
	spellchecker suggester,
	tokenizer Tokenizer,
	fixers ...SegmentFixer,
) *PhraseFixer {
	return &PhraseFixer{
		spellchecker: spellchecker,
		tokenizer:    tokenizer,
		fixers:       fixers,
	}
}

// AddPhrases adds words of the phrases with weight 1.
func (f *PhraseFixer) AddPhrases(phrases ...string) {
	f.AddPhraseWeight(1, phrases...)
}

// AddPhraseWeight adds words of the phrases with the given weight.
func (f *PhraseFixer) AddPhraseWeight(weight uint, phrases ...string) {
	for _, phrase := range phrases {
		for _, token := range f.tokenizer.Tokenize(phrase) {
			f.spellchecker.AddWeight(weight, token.Text)
		}
	}
}

// Fix splits the phrase into segments, finds mistakes in them and passes them through the fixers.
func (f *PhraseFixer) Fix(phrase string, n int, opts ...OptionFunc) PhraseFixResult {
	tokens := f.tokenizer.Tokenize(phrase)
	if len(tokens) == 0 {
		return PhraseFixResult{}
	}

	segments := make([]Segment, 0, len(tokens))

	for _, token := range tokens {
		suggestionResult := f.spellchecker.Suggest(token.Text, n, opts...)

		segments = append(segments, Segment{
			Start:       token.Start,
			End:         token.End,
			Suggestions: suggestionResult.Suggestions,
			Mistakes:    suggestionResult.Mistakes,
		})
	}

	for _, fixer := range f.fixers {
		segments = fixer.Fix(phrase, segments, n, opts...)
	}

	return PhraseFixResult{
		Segments: segments,
	}
}
