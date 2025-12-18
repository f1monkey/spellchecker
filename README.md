# Spellchecker

[![Go Reference](https://pkg.go.dev/badge/github.com/f1monkey/spellchecker.svg)](https://pkg.go.dev/github.com/f1monkey/spellchecker/v3)
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
- average time to fix a single word: ~35 µs
- achieves about 70–74% accuracy on Peter Norvig’s test sets (see [benchmarks](#benchmarks))
- no built-in dictionary — you can provide any custom words, and the spellchecker will only know them

## Installation

```
go get -v github.com/f1monkey/spellchecker/v3
```

## Usage


### Quick start

1. Initialize the spellchecker. You need to pass an alphabet: a set of allowed characters that will be used for indexing and primary word checks. (All other characters will be ignored for these operations.)

```go
	// Create a new instance
	sc, err := spellchecker.New(
		"abcdefghijklmnopqrstuvwxyz1234567890", // allowed symbols, other symbols will be ignored
	)
```

2. Add some words to the dictionary:
	1. from any `io.Reader`:
	```go
		in, _ := os.Open("data/sample.txt")
		sc.AddFrom(in)
	```
	2. Or add words manually:
	```go
		sc.AddMany([]string{"lock", "stock", "and", "two", "smoking"})
		sc.Add("barrels")
	```

3. Use the spellchecker:
	1. Check if a word is correct:
	```go
		result := sc.IsCorrect("stock")
		fmt.Println(result) // true
	```
	2. Suggest corrections:
	```go
		// Find up to 10 suggestions for a word
		matches := sc.Suggest(nil, "rang", 10)
		fmt.Println(matches) // [range, orange]
	```
### Options

### Options

The spellchecker supports customizable options for both searching/suggesting corrections and adding words to the dictionary.

#### Search/Suggestion Options

These options are passed to the `Suggest` method (or to `SuggestWith...` helpers).

- **`SuggestWithMaxErrors(maxErrors int)`**  
  Sets the maximum allowed edit distance (in "bits") between the input word and dictionary candidates.  
  - Deletion: 1 bit (e.g., "proble" → "problem")  
  - Insertion: 1 bit (e.g., "problemm" → "problem")  
  - Substitution: 2 bits (e.g., "problam" → "problem")  
  - Transposition: 0 bits (e.g., "problme" → "problem")  

  Default: `2`.
  Increasing this value beyond 2 is not recommended as it can significantly degrade performance.

- **`SuggestWithFilterFunc(f FilterFunc)`**  
  Replaces the default scoring/filtering function with a custom one.  
  The function receives:
  - `src`: runes of the input word
  - `candidate`: runes of the dictionary word
  - `count`: frequency count of the candidate in the dictionary

  It must return:
  - a `float64` score (higher = better suggestion)
  - a `bool` indicating whether the candidate should be kept

  The default filter uses Levenshtein distance (with costs: insert/delete=1, substitute=1, transpose=1), filters out candidates exceeding `maxErrors`, and boosts score based on word frequency and shared prefix/suffix length.

Example usage:
```go
matches := sc.Suggest(
	"rang",
	10,
	spellchecker.SuggestWithMaxErrors(1),
	spellchecker.SuggestWithFilterFunc(myCustomFilter),
)
```

#### Add Options
These options are passed to `Add`, `AddMany`, or `AddFrom`.

- **`AddWithWeight(weight uint)`**
  Sets the frequency weight for added word(s). Higher weight increases the chance that the word will appear higher in suggestion results.
  Default: 1.
- **`AddWithSplitter(splitter bufio.SplitFunc)`**
  Customizes how AddFrom(reader) splits the input stream into words.

  The default splitter:
    - Uses bufio.ScanWords as base
    - Converts to lowercase
    - Keeps only sequences matching [-\pL]+ (letters and hyphens)

Example:
```go
sc.AddFrom(
	file,
	spellchecker.AddWithWeight(10),          // these words are very common
	spellchecker.AddWithSplitter(customSplitter),
)

sc.AddMany([]string{"hello", "world"},
	spellchecker.AddWithWeight(5),
)
```

### Save/load

```go
	sc, err := spellchecker.New("abc")

	// Save data to any io.Writer
	out, err := os.Create("data/out.bin")
	if err != nil {
		panic(err)
	}
	sc.Save(out)

	// Load data back from io.Reader
	in, err = os.Open("data/out.bin")
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
Running tool: /usr/bin/go test -benchmem -run=^$ -bench ^Benchmark_Norvig1$ github.com/f1monkey/spellchecker -count=1

goos: linux
goarch: amd64
pkg: github.com/f1monkey/spellchecker
cpu: 13th Gen Intel(R) Core(TM) i9-13980HX
Benchmark_Norvig1-32    	     357	   3305052 ns/op	        74.44 success_percent	       201.0 success_words	       270.0 total_words	  768899 B/op	   13302 allocs/op
PASS
ok  	github.com/f1monkey/spellchecker	3.801s
```

#### [Test set 2](http://norvig.com/spell-testset2.txt):

```
Running tool: /usr/bin/go test -benchmem -run=^$ -bench ^Benchmark_Norvig2$ github.com/f1monkey/spellchecker -count=1

goos: linux
goarch: amd64
pkg: github.com/f1monkey/spellchecker
cpu: 13th Gen Intel(R) Core(TM) i9-13980HX
Benchmark_Norvig2-32    	     236	   5257185 ns/op	        71.25 success_percent	       285.0 success_words	       400.0 total_words	 1201260 B/op	   19346 allocs/op
PASS
ok  	github.com/f1monkey/spellchecker	4.350s
```
