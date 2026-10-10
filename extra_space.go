package spellchecker

import "strings"

type wordChecker interface {
	IsCorrect(word string) bool
}

// ExtraSpaceCorrector joins words split by extra spaces or punctuation.
type ExtraSpaceCorrector struct {
	checker wordChecker
	maxJoin int
}

// NewExtraSpaceCorrector creates an ExtraSpaceCorrector.
// maxJoin is the max number of segments joined into one word.
func NewExtraSpaceCorrector(checker wordChecker, maxJoin int) *ExtraSpaceCorrector {
	return &ExtraSpaceCorrector{
		checker: checker,
		maxJoin: maxJoin,
	}
}

// Correct replaces neighbor segments with a correct word joined from them without the text between.
// Segments are joined only if at least one of them has a mistake, so correct words stay separate.
func (c *ExtraSpaceCorrector) Correct(phrase Phrase, segments []Segment, _ ...Option) []Segment {
	result := make([]Segment, 0, len(segments))

	for i := 0; i < len(segments); {
		last, word := c.join(phrase.Text, segments, i)
		if last == i {
			result = append(result, segments[i])
			i++

			continue
		}

		result = append(result, Segment{
			Text:        phrase.Text[segments[i].Start:segments[last].End],
			Start:       segments[i].Start,
			End:         segments[last].End,
			Suggestions: []Suggestion{{Value: word}},
			Mistakes:    MistakeExtraSpace,
		})
		i = last + 1
	}

	return result
}

// join finds the longest correct word made of segments[first:last+1].
// It returns first if there is no such word.
func (c *ExtraSpaceCorrector) join(phrase string, segments []Segment, first int) (int, string) {
	end := min(first+c.maxJoin, len(segments)) - 1

	withMistake := first
	for withMistake <= end && segments[withMistake].IsCorrect() {
		withMistake++
	}

	if end == first || withMistake > end {
		return first, ""
	}

	var b strings.Builder

	lengths := make([]int, end-first+1)

	for i := first; i <= end; i++ {
		b.WriteString(phrase[segments[i].Start:segments[i].End])
		lengths[i-first] = b.Len()
	}

	joined := b.String()
	for last := end; last > first && last >= withMistake; last-- {
		word := joined[:lengths[last-first]]

		if c.checker.IsCorrect(word) {
			return last, word
		}
	}

	return first, ""
}
