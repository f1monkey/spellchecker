package spellchecker

import (
	"bufio"
	"bytes"
	"math"
	"regexp"

	"github.com/agext/levenshtein"
)

const DefaultMaxErrors = 2

type FilterFunc func(src, candidate []rune, count uint) (float64, bool)

type SearchOptions struct {
	// MaxErrors — the maximum allowed difference in bits
	// between the "search word" and a "dictionary word".
	// - deletion is a 1-bit change (proble → problem)
	// - insertion is a 1-bit change (problemm → problem)
	// - substitution is a 2-bit change (problam → problem)
	// - transposition is a 0-bit change (problme → problem)
	//
	// It is not recommended to set this value greater than 2,
	// as it can significantly affect performance.
	MaxErrors int

	// FilterFunc compares the source word with a candidate word.
	// It returns the candidate's score and a boolean flag.
	// If the flag is false, the candidate will be completely filtered out.
	FilterFunc FilterFunc
}

var defaultSearchOptions = &SearchOptions{
	MaxErrors:  DefaultMaxErrors,
	FilterFunc: defaultFilterFunc(DefaultMaxErrors),
}

type AddOptions struct {
	Weight uint
	// Splitter is a splitter func for AddFrom() reader
	Splitter bufio.SplitFunc
}

var defaultAddOptions = &AddOptions{
	Weight:   1,
	Splitter: defaultSplitter,
}

var wordSymbols = regexp.MustCompile(`[-\pL]+`)

func defaultSplitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanWords(data, atEOF)
	if err != nil {
		return
	}
	token = bytes.ToLower(token)

	return advance, wordSymbols.Find(token), nil
}

func defaultFilterFunc(maxErrors int) FilterFunc {
	return func(src, candidate []rune, count uint) (float64, bool) {
		distance, prefixLen, suffixLen := levenshtein.Calculate(src, candidate, 0, 1, 1, 1)
		if distance > maxErrors {
			return 0, false
		}

		mult := math.Log1p(float64(count)) * math.Pow(1.5, float64(prefixLen+suffixLen))

		return 1 / (1 + float64(distance*distance)) * mult, true
	}
}

func applyDefaults(opts *SearchOptions) *SearchOptions {
	if opts == nil {
		opts = defaultSearchOptions
	} else {
		if opts.MaxErrors == 0 {
			opts.MaxErrors = DefaultMaxErrors
		}
		if opts.FilterFunc == nil {
			opts.FilterFunc = defaultFilterFunc(opts.MaxErrors)
		}
	}

	return opts
}
