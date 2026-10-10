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

`PhraseFixer` splits a phrase into segments with a tokenizer and checks each word. Correctors fix more mistakes, they are applied in the given order:

- `NewExtraSpaceCorrector(maxJoin)` joins words split by spaces or punctuation: "hel lo" → "hello".
- `NewLayoutCorrector(spellchecker.QwertyRuEn())` fixes words typed in a wrong keyboard layout: "ghbdtn" → "привет".

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

You can also implement the `Corrector` interface. `Correct` receives a `CorrectInput` with the phrase and the spellchecker and tokenizer of `PhraseFixer`.

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
