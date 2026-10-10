# Spellchecker

[![Go Reference](https://pkg.go.dev/badge/github.com/f1monkey/spellchecker.svg)](https://pkg.go.dev/badge/github.com/f1monkey/spellchecker/v4)
[![CI](https://github.com/f1monkey/spellchecker/actions/workflows/test.yaml/badge.svg)](https://github.com/f1monkey/spellchecker/actions/workflows/test.yaml)

Yet another spellchecker written in go.

- [Spellchecker](#spellchecker)
  - [Features:](#features)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Quick start](#quick-start)
    - [Options](#options)
    - [Phrases](#phrases)
      - [Mistakes](#mistakes)
      - [ExtraSpaceCorrector](#extraspacecorrector)
      - [LayoutCorrector](#layoutcorrector)
      - [Order of correctors](#order-of-correctors)
      - [Custom correctors](#custom-correctors)
    - [Tokenizers](#tokenizers)
    - [Save/load](#saveload)
  - [Benchmarks](#benchmarks)
    - [Test set 1:](#test-set-1)
    - [Test set 2:](#test-set-2)

## Features:

- very compact database: ~1 MB for 30,000 unique words
- average time to fix a single word: ~12 µs
- achieves about 70–74% accuracy on Peter Norvig’s test sets (see [benchmarks](#benchmarks))
- no built-in dictionary — you can provide any custom words, and the spellchecker will only know them

## Installation

```
go get -v github.com/f1monkey/spellchecker/v4
```

## Usage

### Quick start

1. Initialize the spellchecker. Pass one or more alphabets: sets of allowed characters used for indexing and lookup. Characters outside the alphabet are ignored for these operations.

```go
	sc, err := spellchecker.New(
		spellchecker.EN, spellchecker.Numbers, // or a custom string like "abcdefghijklmnopqrstuvwxyz1234567890"
	)
```

2. Add words to the dictionary:
   ```go
   	sc.Add("lock", "stock", "barrels")
   	sc.AddWeight(5, "hello", "world") // higher weight ranks the word higher in suggestions
   ```

3. Use the spellchecker:
   1. Check if a word is correct:

   ```go
   	result := sc.IsCorrect("stock")
   	fmt.Println(result) // true
   ```

   2. Suggest corrections:

   ```go
   	result := sc.Suggest("rang")
   	fmt.Println(result.Suggestions) // [{range ...} {orange ...}]
   ```

### Options

These options are passed to `Suggest` and `PhraseFixer.Fix`.

- **`WithMaxSuggestions(n int)`**
  Sets the maximum number of suggestions for a word.

  Default: `10`.

- **`WithMaxErrors(maxErrors int)`**
  Sets the maximum allowed difference in bits between the input word and dictionary candidates.
  - Deletion: 1 bit (e.g., "proble" → "problem")
  - Insertion: 1 bit (e.g., "problemm" → "problem")
  - Substitution: 2 bits (e.g., "problam" → "problem")
  - Transposition: 0 bits (e.g., "problme" → "problem")

  The same value is passed to the scoring function as the maximum allowed edit distance.

  Default: `2`.
  Increasing this value beyond 2 is not recommended as it can significantly degrade performance.

- **`WithScoringFunc(f ScoringFunc)`**
  Replaces the default scoring/filtering function with a custom one.  
  The function receives:
  - `src`: runes of the input word
  - `candidate`: runes of the dictionary word
  - `count`: frequency count of the candidate in the dictionary
  - `maxErrors`: the value set by `WithMaxErrors`

  It must return:
  - a `float64` score (higher = better suggestion)
  - a `bool` indicating whether the candidate should be kept

  The default scoring uses Levenshtein distance (insertion, deletion and substitution cost 1 each; a transposition of adjacent letters counts as 2 edits). It scorings out candidates whose distance exceeds `maxErrors`, and boosts the score based on word frequency and shared prefix/suffix length.

Example usage:

```go
result := sc.Suggest(
	"rang",
	spellchecker.WithMaxSuggestions(5),
	spellchecker.WithMaxErrors(1),
	spellchecker.WithScoringFunc(myCustomScoring),
)
```

### Phrases

`PhraseFixer` fixes whole phrases:

1. The tokenizer splits the phrase into words. Each word becomes a `Segment`.
2. Each word is checked with `Suggest`: the segment gets its `Mistakes` and `Suggestions`.
3. The segments are passed through the correctors in the given order. Each corrector gets the result of the previous one and may replace any segments with new ones.

The result is a list of segments in phrase order. Text between segments (spaces, punctuation) is not in the result, it stays as is.

```go
	sc, err := spellchecker.New(spellchecker.EN, spellchecker.RU)
	if err != nil {
		panic(err)
	}

	f := spellchecker.NewPhraseFixer(sc, spellchecker.NewStandardTokenizer(),
		spellchecker.NewExtraSpaceCorrector(3),
		spellchecker.NewLayoutCorrector(spellchecker.QwertyRuEn()),
	)

	f.Add("hello world привет") // adds words of the phrases to the spellchecker

	phrase := "hel lo ghbdtn"
	result := f.Fix(phrase, spellchecker.WithMaxSuggestions(3))
	fmt.Println(result.Apply(phrase)) // "hello привет"

	for _, s := range result.Segments {
		// s.Start and s.End are byte offsets in the phrase, s.Text is phrase[s.Start:s.End]
		fmt.Println(s.Text, s.Mistakes, s.Suggestions)
		// hel lo extra_space [{hello 0}]
		// ghbdtn layout [{привет 0}]
	}
```

`Apply` replaces each wrong segment with its first suggestion. Segments without suggestions (unknown words) are kept.

#### Mistakes

`Segment.Mistakes` is a bit set, check it with `Has`: `s.Mistakes.Has(spellchecker.MistakeTypo)`.

| Mistake | Set by | Meaning |
|---|---|---|
| `NoMistake` | `Suggest` | the word is in the dictionary |
| `MistakeTypo` | `Suggest` | the word is not in the dictionary, `Suggestions` has similar words |
| `MistakeUnknownWord` | `Suggest` | the word is not in the dictionary and there are no similar words |
| `MistakeLayout` | `LayoutCorrector` | the text is typed in a wrong keyboard layout |
| `MistakeExtraSpace` | `ExtraSpaceCorrector` | one word is split by spaces or punctuation |

Correctors may combine bits: a word typed in a wrong layout with a typo is `MistakeLayout | MistakeTypo`.

#### ExtraSpaceCorrector

`NewExtraSpaceCorrector(maxJoin)` joins words split by extra spaces or punctuation: "hel lo" → "hello", "wi th,out" → "without".

- It tries to join up to `maxJoin` neighbor segments. The text between them is dropped, so "hel, lo" is joined into "hello" too.
- At least one of the joined segments must have a mistake. Correct words are never joined: "in to" stays "in to", even if "into" is in the dictionary.
- The joined word must be in the dictionary. A joined word with a typo is not used.
- The longest join wins: "wi th out" → "without", not "with" + "out".
- The joined segments are replaced with one segment: `Text` is the original text ("hel lo"), `Suggestions` is the joined word, `Mistakes` is `MistakeExtraSpace`.

Words are checked with `IsCorrect` only, so this corrector is cheap: it does no fuzzy search.

Words are joined through any punctuation, also through the end of a sentence: "...hel. lo..." becomes "...hello...", if "hel" or "lo" is wrong.

#### LayoutCorrector

`NewLayoutCorrector(replacer)` fixes text typed in a wrong keyboard layout: "ghbdtn" → "привет", "руддщ" → "hello". `spellchecker.QwertyRuEn()` switches between US QWERTY and Russian ЙЦУКЕН in both directions. You can make your own layout with `spellchecker.LayoutReplacer`: a map from a rune to a rune.

- The layout is switched for whole whitespace-separated chunks, not for single segments. Punctuation keys may be letters in the other layout: "j,]tv" is "объем", but the standard tokenizer would split it into "j" and "tv".
- A chunk is switched only if at least one of its segments has a mistake.
- The switched chunk is split into words by the `PhraseFixer` tokenizer, and each word is checked with `Suggest`. The switch is "all or nothing":
  - if any switched word is unknown, the chunk is kept as is;
  - if the original chunk has an unknown word, the switch is used when all switched words are found: correct words or typos ("ghbdtnn" → "приветт" → "привет");
  - if the original chunk has only typos, the switch is used only when all switched words are correct. Otherwise the original typo is better than a typo in another layout.
- The switched words become new segments with `MistakeLayout`. A correct switched word has itself as the only suggestion. A switched word with a typo also has `MistakeTypo` and suggestions for the switched word.
- Text between words that changes in the other layout gets its own segment, so that `Apply` switches it too: "руддщбцщкдв" → "hello", ",", "world".
- Chunks with a segment that crosses a whitespace are skipped. Such segments are made by other correctors, for example by `ExtraSpaceCorrector`.

Each switched word costs one `Suggest` call, so this corrector is more expensive than `ExtraSpaceCorrector`.

#### Order of correctors

Correctors work with the original text of the phrase, so the order matters. Put `ExtraSpaceCorrector` before `LayoutCorrector`:

- `ExtraSpaceCorrector` joins the original text. If `LayoutCorrector` runs first, the switched segments still have the text in the old layout, and they could be joined into a wrong word.
- `LayoutCorrector` skips the joined segments, because they cross whitespace.

So "ghb dtn" (wrong layout and an extra space) is not fixed by both correctors together.

#### Custom correctors

Implement the `Corrector` interface. `Correct` receives a `CorrectInput` with the phrase and the spellchecker and the tokenizer of `PhraseFixer`, and the segments from the previous corrector. It must return segments in phrase order, which don't overlap, with `Text` equal to `Phrase[Start:End]`. Don't modify the input slice, return a new one.

For example, a corrector that ignores mistakes in short words:

```go
type shortWordsCorrector struct{}

func (shortWordsCorrector) Correct(
	in spellchecker.CorrectInput,
	segments []spellchecker.Segment,
	_ ...spellchecker.Option,
) []spellchecker.Segment {
	result := make([]spellchecker.Segment, 0, len(segments))

	for _, s := range segments {
		if utf8.RuneCountInString(s.Text) < 3 {
			s.Suggestions, s.Mistakes = nil, spellchecker.NoMistake
		}

		result = append(result, s)
	}

	return result
}
```

### Tokenizers

Tokenizers are used by `PhraseFixer`.

- `NewWhitespaceTokenizer` splits on Unicode whitespace (like Elasticsearch `whitespace`).
- `NewStandardTokenizer` approximates Elasticsearch `standard`: it keeps letters, digits, underscores and in-word apostrophes, and splits on hyphens and other punctuation.
- `NewRegexpTokenizer` - splits strings using the provided regular expression
- You can also implement the `Tokenizer` interface

### Save/load

```go
	sc, err := spellchecker.New(spellchecker.EN)

	// Save data to any io.Writer
	out, err := os.Create("data/out.bin")
	if err != nil {
		panic(err)
	}
	sc.Save(out)

	// Load data back from io.Reader
	in, err := os.Open("data/out.bin")
	if err != nil {
		panic(err)
	}
	sc, err = spellchecker.Load(in)
	if err != nil {
		panic(err)
	}
```

## Benchmarks

Tests are based on data from [Peter Norvig's article about spelling correction](http://norvig.com/spell-correct.html)

#### [Test set 1](http://norvig.com/spell-testset1.txt):

```
Running tool: /usr/bin/go test -benchmem -run=^$ -bench ^Benchmark_Norvig1$ github.com/f1monkey/spellchecker/v4 -count=1

goos: linux
goarch: amd64
pkg: github.com/f1monkey/spellchecker/v4
cpu: AMD Ryzen 9 9950X3D 16-Core Processor
Benchmark_Norvig1-32                             	     363	   3234033 ns/op	        74.07 success_percent	       200.0 success_words	       270.0 total_words	  119670 B/op	    2314 allocs/op
PASS
ok  	github.com/f1monkey/spellchecker/v4	3.565s
```

#### [Test set 2](http://norvig.com/spell-testset2.txt):

```
Running tool: /usr/bin/go test -benchmem -run=^$ -bench ^Benchmark_Norvig2$ github.com/f1monkey/spellchecker/v4 -count=1

goos: linux
goarch: amd64
pkg: github.com/f1monkey/spellchecker/v4
cpu: AMD Ryzen 9 9950X3D 16-Core Processor
Benchmark_Norvig2-32                             	     258	   4379073 ns/op	        71.25 success_percent	       285.0 success_words	       400.0 total_words	  170356 B/op	    3063 allocs/op
PASS
ok  	github.com/f1monkey/spellchecker/v4	3.844s
```
