package spellchecker

import (
	"strings"
	"unicode/utf8"
)

// LayoutReplacer replaces runes by the map.
type LayoutReplacer map[rune]rune

func (r LayoutReplacer) Replace(s string) string {
	return strings.Map(func(c rune) rune {
		if mapped, ok := r[c]; ok {
			return mapped
		}

		return c
	}, s)
}

// QwertyRuEn switches text between US QWERTY and Russian ЙЦУКЕН layouts.
// Russian punctuation is not mapped, because it is on other keys than the same English characters.
var QwertyRuEn = LayoutReplacer(map[rune]rune{
	// en to ru
	'`': 'ё', 'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н', 'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з', '[': 'х', ']': 'ъ',
	'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р', 'j': 'о', 'k': 'л', 'l': 'д', ';': 'ж', '\'': 'э',
	'z': 'я', 'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь', ',': 'б', '.': 'ю',
	'~': 'Ё', 'Q': 'Й', 'W': 'Ц', 'E': 'У', 'R': 'К', 'T': 'Е', 'Y': 'Н', 'U': 'Г', 'I': 'Ш', 'O': 'Щ', 'P': 'З', '{': 'Х', '}': 'Ъ',
	'A': 'Ф', 'S': 'Ы', 'D': 'В', 'F': 'А', 'G': 'П', 'H': 'Р', 'J': 'О', 'K': 'Л', 'L': 'Д', ':': 'Ж', '"': 'Э',
	'Z': 'Я', 'X': 'Ч', 'C': 'С', 'V': 'М', 'B': 'И', 'N': 'Т', 'M': 'Ь', '<': 'Б', '>': 'Ю',

	// ru to en
	'ё': '`', 'й': 'q', 'ц': 'w', 'у': 'e', 'к': 'r', 'е': 't', 'н': 'y', 'г': 'u', 'ш': 'i', 'щ': 'o', 'з': 'p', 'х': '[', 'ъ': ']',
	'ф': 'a', 'ы': 's', 'в': 'd', 'а': 'f', 'п': 'g', 'р': 'h', 'о': 'j', 'л': 'k', 'д': 'l', 'ж': ';', 'э': '\'',
	'я': 'z', 'ч': 'x', 'с': 'c', 'м': 'v', 'и': 'b', 'т': 'n', 'ь': 'm', 'б': ',', 'ю': '.',
	'Ё': '~', 'Й': 'Q', 'Ц': 'W', 'У': 'E', 'К': 'R', 'Е': 'T', 'Н': 'Y', 'Г': 'U', 'Ш': 'I', 'Щ': 'O', 'З': 'P', 'Х': '{', 'Ъ': '}',
	'Ф': 'A', 'Ы': 'S', 'В': 'D', 'А': 'F', 'П': 'G', 'Р': 'H', 'О': 'J', 'Л': 'K', 'Д': 'L', 'Ж': ':', 'Э': '"',
	'Я': 'Z', 'Ч': 'X', 'С': 'C', 'М': 'V', 'И': 'B', 'Т': 'N', 'Ь': 'M', 'Б': '<', 'Ю': '>',
})

type replacer interface {
	Replace(s string) string
}

// LayoutCorrector corrects words typed in a wrong keyboard layout.
type LayoutCorrector struct {
	replacer            replacer
	suggester           suggester
	tokenizer           Tokenizer
	whitespaceTokenizer Tokenizer
}

// NewLayoutCorrector creates a LayoutCorrector. The replacer must map each rune to one rune.
// The tokenizer must be the same as in PhraseFixer.
func NewLayoutCorrector(replacer replacer, suggester suggester, tokenizer Tokenizer) *LayoutCorrector {
	return &LayoutCorrector{
		replacer:            replacer,
		suggester:           suggester,
		tokenizer:           tokenizer,
		whitespaceTokenizer: NewWhitespaceTokenizer(),
	}
}

// Correct replaces segments typed in a wrong layout.
// The layout is switched for whole whitespace-separated chunks, because other separators
// may be letters in another layout.
func (f *LayoutCorrector) Correct(phrase string, segments []Segment, maxSuggestions int, opts ...OptionFunc) []Segment {
	result := make([]Segment, 0, len(segments))
	i := 0

	for _, chunk := range f.whitespaceTokenizer.Tokenize(phrase) {
		first := i
		mistakes := NoMistake
		inside := true

		for ; i < len(segments) && segments[i].Start < chunk.End; i++ {
			mistakes |= segments[i].Mistakes
			inside = inside && segments[i].Start >= chunk.Start && segments[i].End <= chunk.End
		}

		if mistakes != NoMistake && inside {
			if fixed := f.correctChunk(phrase, chunk.Start, chunk.End, mistakes, maxSuggestions, opts...); fixed != nil {
				result = append(result, fixed...)

				continue
			}
		}

		result = append(result, segments[first:i]...)
	}

	return append(result, segments[i:]...)
}

// correctChunk switches the layout of phrase[start:end] and returns its segments, or nil if it is not better.
// mistakes are mistakes of the original segments of this text.
func (f *LayoutCorrector) correctChunk(
	phrase string,
	start, end int,
	mistakes Mistake,
	maxSuggestions int,
	opts ...OptionFunc,
) []Segment {
	text := phrase[start:end]

	replaced := f.replacer.Replace(text)
	if replaced == text {
		return nil
	}

	tokens := f.tokenizer.Tokenize(replaced)
	if len(tokens) == 0 || !toSourceOffsets(text, replaced, tokens) {
		return nil
	}

	result := make([]Segment, 0, len(tokens))

	// appendGap adds a segment for text between tokens, otherwise it stays in the original layout.
	appendGap := func(gapStart, gapEnd int) {
		if gapStart == gapEnd {
			return
		}

		gap := text[gapStart:gapEnd]

		replacedGap := f.replacer.Replace(gap)
		if replacedGap == gap {
			return
		}

		result = append(result, Segment{
			Start:       start + gapStart,
			End:         start + gapEnd,
			Suggestions: []Suggestion{{Value: replacedGap}},
			Mistakes:    MistakeLayout,
		})
	}

	allCorrect := true
	prevEnd := 0

	for _, token := range tokens {
		suggestions := f.suggester.Suggest(token.Text, maxSuggestions, opts...)
		if suggestions.Mistakes.Has(MistakeUnknownWord) {
			return nil
		}

		appendGap(prevEnd, token.Start)
		prevEnd = token.End

		newSegment := Segment{
			Start:       start + token.Start,
			End:         start + token.End,
			Suggestions: suggestions.Suggestions,
			Mistakes:    suggestions.Mistakes | MistakeLayout,
		}

		if suggestions.IsCorrect() {
			newSegment.Suggestions = []Suggestion{{Value: token.Text}}
		} else {
			allCorrect = false
		}

		result = append(result, newSegment)
	}

	if !mistakes.Has(MistakeUnknownWord) && !allCorrect {
		return nil
	}

	appendGap(prevEnd, len(text))

	return result
}

// toSourceOffsets converts token offsets in replaced into offsets in source.
func toSourceOffsets(source, replaced string, tokens []Token) bool {
	if utf8.RuneCountInString(source) != utf8.RuneCountInString(replaced) {
		return false
	}

	i, j := 0, 0
	advanceTo := func(offset int) int {
		for i < offset {
			_, n := utf8.DecodeRuneInString(replaced[i:])
			_, m := utf8.DecodeRuneInString(source[j:])
			i += n
			j += m
		}

		return j
	}

	for k := range tokens {
		tokens[k].Start = advanceTo(tokens[k].Start)
		tokens[k].End = advanceTo(tokens[k].End)
	}

	return true
}
