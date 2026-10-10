package spellchecker

import "strings"

type suggester interface {
	Suggest(word string, opts ...Option) SuggestionResult
	IsCorrect(word string) bool
	AddWeight(weight uint, words ...string)
}

// Segment is a part of the phrase: a correct word or a fragment to fix.
// Start and End are byte offsets in the phrase.
type Segment struct {
	// Text is the original text of the segment.
	Text  string
	Start int
	End   int
	// Suggestions are fix candidates, best first.
	Suggestions []Suggestion
	// Mistakes are mistakes found in the segment.
	Mistakes Mistake
}

// IsCorrect reports whether the segment has no mistakes.
func (s Segment) IsCorrect() bool { return s.Mistakes == NoMistake }

// PhraseFixResult is the result of PhraseFixer.Fix.
type PhraseFixResult struct {
	// Segments are in phrase order.
	Segments []Segment
}

// Apply returns the phrase with each wrong segment replaced by its best suggestion.
func (r PhraseFixResult) Apply(phrase string) string {
	var b strings.Builder

	prev := 0

	for _, s := range r.Segments {
		if s.IsCorrect() || len(s.Suggestions) == 0 {
			continue
		}

		if b.Len() == 0 {
			b.Grow(len(phrase))
		}

		b.WriteString(phrase[prev:s.Start])
		b.WriteString(s.Suggestions[0].Value)
		prev = s.End
	}

	if prev == 0 {
		return phrase
	}

	b.WriteString(phrase[prev:])

	return b.String()
}

// CorrectInput is the phrase with the spellchecker and the tokenizer of PhraseFixer.
type CorrectInput struct {
	Phrase       string
	Tokenizer    Tokenizer
	Spellchecker suggester
}

// Corrector corrects segments of the phrase found by PhraseFixer.
type Corrector interface {
	Correct(in CorrectInput, segments []Segment, opts ...Option) []Segment
}

// PhraseFixer fixes mistakes in phrases.
type PhraseFixer struct {
	spellchecker suggester
	tokenizer    Tokenizer
	correctors   []Corrector
}

// NewPhraseFixer creates a PhraseFixer. Correctors are applied in the given order.
func NewPhraseFixer(
	spellchecker suggester,
	tokenizer Tokenizer,
	correctors ...Corrector,
) *PhraseFixer {
	return &PhraseFixer{
		spellchecker: spellchecker,
		tokenizer:    tokenizer,
		correctors:   correctors,
	}
}

// Add adds words of the phrases with weight 1.
func (f *PhraseFixer) Add(phrases ...string) {
	f.AddWeight(1, phrases...)
}

// AddWeight adds words of the phrases with the given weight.
func (f *PhraseFixer) AddWeight(weight uint, phrases ...string) {
	for _, phrase := range phrases {
		for _, token := range f.tokenizer.Tokenize(phrase) {
			f.spellchecker.AddWeight(weight, token.Text)
		}
	}
}

// Fix splits the phrase into segments, finds mistakes in them and passes them through the correctors.
func (f *PhraseFixer) Fix(phrase string, opts ...Option) PhraseFixResult {
	tokens := f.tokenizer.Tokenize(phrase)
	if len(tokens) == 0 {
		return PhraseFixResult{}
	}

	segments := make([]Segment, 0, len(tokens))

	for _, token := range tokens {
		suggestionResult := f.spellchecker.Suggest(token.Text, opts...)

		segments = append(segments, Segment{
			Text:        token.Text,
			Start:       token.Start,
			End:         token.End,
			Suggestions: suggestionResult.Suggestions,
			Mistakes:    suggestionResult.Mistakes,
		})
	}

	in := CorrectInput{Phrase: phrase, Tokenizer: f.tokenizer, Spellchecker: f.spellchecker}
	for _, corrector := range f.correctors {
		segments = corrector.Correct(in, segments, opts...)
	}

	return PhraseFixResult{
		Segments: segments,
	}
}
