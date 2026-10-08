# Spellchecker

[![Go Reference](https://pkg.go.dev/badge/github.com/f1monkey/spellchecker.svg)](https://pkg.go.dev/badge/github.com/f1monkey/spellchecker/v4)
[![CI](https://github.com/f1monkey/spellchecker/actions/workflows/test.yaml/badge.svg)](https://github.com/f1monkey/spellchecker/actions/workflows/test.yaml)

Yet another spellchecker written in go.

- [Spellchecker](#spellchecker)
  - [Features:](#features)
  - [Installation](#installation)
  - [Usage](#usage)
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

1. Initialize the spellchecker. Pass a tokenizer and one or more alphabets: sets of allowed characters used for indexing and lookup. Characters outside the alphabet are ignored for these operations.

```go
	sc, err := spellchecker.New(
		spellchecker.NewWhitespaceTokenizer(), // or NewStandardTokenizer()/NewRegexpTokenizer()
		spellchecker.EN, spellchecker.Numbers, // or a custom string like "abcdefghijklmnopqrstuvwxyz1234567890"
	)
```

`NewWhitespaceTokenizer` splits on Unicode whitespace (like Elasticsearch `whitespace`).
`NewStandardTokenizer` approximates Elasticsearch `standard`: it keeps letters, digits, underscores and in-word apostrophes, and splits on hyphens and other punctuation. You can also implement `Tokenizer` inteface or use `NewRegexpTokenizer`.

2. Add words to the dictionary:
   1. Whole words (no splitting):

   ```go
   	sc.Add("lock", "stock", "barrels")
   	sc.AddWeight(5, "hello", "world") // higher weight ranks the word higher in suggestions
   ```

   2. Phrases — tokenized with the tokenizer from `New`:

   ```go
   	sc.AddPhrases("lock stock and two smoking barrels")
   	sc.AddPhraseWeight(10, "very common phrase")
   ```

3. Use the spellchecker:
   1. Check if a word is correct:

   ```go
   	result := sc.IsCorrect("stock")
   	fmt.Println(result) // true
   ```

   2. Suggest corrections:

   ```go
   	result := sc.Suggest("rang", 10)
   	fmt.Println(result.Suggestions) // [{range ...} {orange ...}]
   ```

### Options

These options are passed to `Suggest`.

- **`WithMaxErrors(maxErrors int)`**
  Sets the maximum allowed difference in bits between the input word and dictionary candidates.
  - Deletion: 1 bit (e.g., "proble" → "problem")
  - Insertion: 1 bit (e.g., "problemm" → "problem")
  - Substitution: 2 bits (e.g., "problam" → "problem")
  - Transposition: 0 bits (e.g., "problme" → "problem")

  The same value is passed to the filter function as the maximum allowed edit distance.

  Default: `2`.
  Increasing this value beyond 2 is not recommended as it can significantly degrade performance.

- **`WithFilterFunc(f FilterFunc)`**
  Replaces the default scoring/filtering function with a custom one.  
  The function receives:
  - `src`: runes of the input word
  - `candidate`: runes of the dictionary word
  - `count`: frequency count of the candidate in the dictionary
  - `maxErrors`: the value set by `WithMaxErrors`

  It must return:
  - a `float64` score (higher = better suggestion)
  - a `bool` indicating whether the candidate should be kept

  The default filter uses Levenshtein distance (insertion, deletion and substitution cost 1 each; a transposition of adjacent letters counts as 2 edits). It filters out candidates whose distance exceeds `maxErrors`, and boosts the score based on word frequency and shared prefix/suffix length.

Example usage:

```go
result := sc.Suggest(
	"rang",
	10,
	spellchecker.WithMaxErrors(1),
	spellchecker.WithFilterFunc(myCustomFilter),
)
```

### Save/load

```go
	tok := spellchecker.NewWhitespaceTokenizer()
	sc, err := spellchecker.New(tok, "abc")

	// Save data to any io.Writer
	out, err := os.Create("data/out.bin")
	if err != nil {
		panic(err)
	}
	sc.Save(out)

	// Load data back from io.Reader (pass a tokenizer for AddPhrases after load)
	in, err := os.Open("data/out.bin")
	if err != nil {
		panic(err)
	}
	sc, err = spellchecker.Load(in, tok)
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
Benchmark_Norvig1-32    	     360	   3327587 ns/op	        74.07 success_percent	       200.0 success_words	       270.0 total_words	  119688 B/op	    2314 allocs/op
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
Benchmark_Norvig2-32    	     256	   4699043 ns/op	        71.00 success_percent	       284.0 success_words	       400.0 total_words	  170442 B/op	    3062 allocs/op
PASS
ok  	github.com/f1monkey/spellchecker/v4	3.844s
```
