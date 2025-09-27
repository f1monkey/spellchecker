# Spellchecker

[![Go Reference](https://pkg.go.dev/badge/github.com/f1monkey/spellchecker.svg)](https://pkg.go.dev/github.com/f1monkey/spellchecker)
[![CI](https://github.com/f1monkey/spellchecker/actions/workflows/test.yml/badge.svg)](https://github.com/f1monkey/spellchecker/actions/workflows/test.yml)

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

## Installation

```
go get github.com/f1monkey/spellchecker/v2@latest
```

## Usage


### Quick start

```go

func main() {
	// Create a new instance
	sc, err := spellchecker.New(
		"abcdefghijklmnopqrstuvwxyz1234567890", // allowed symbols, other symbols will be ignored
	)
	if err != nil {
		panic(err)
	}

	// The weight increases the likelihood that the word will be chosen as a correction.
	weight := uint(1)

	// Load data from any io.Reader
	in, err := os.Open("data/sample.txt")
	if err != nil {
		panic(err)
	}

	sc.AddFrom(&spellchecker.AddOptions{Weight: weight}, in)
	// OR
	sc.AddFrom(nil, in)

	// Add words manually
	sc.Add(nil, "lock", "stock", "and", "two", "smoking", "barrels")

	// Check if a word is valid
	result := sc.IsCorrect("coffee")
	fmt.Println(result) // true

	// Correct a single word
	fixed, isCorrect := sc.Fix(nil, "awepon")
	fmt.Println(isCorrect) // false
	fmt.Println(fixed) // weapon

	// Find up to 10 suggestions for a word
	matches := sc.Suggest(nil, "rang", 10)
	fmt.Println(matches) // [range, orange]

	if len(os.Args) < 2 {
		log.Fatal("dict path must be provided")
	}
```

### Options

See [options.go](./options.go) for the list of available options.

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

### Custom score function

You can provide a custom scoring function if needed:

```go
	var fn spellchecker.FilterFunc = func(src, candidate []rune, cnt int) (float64, bool) {
		// you can calculate Levenshtein distance here (see defaultFilterFunc in options.go for example)

		return 1.0, true // constant score
	}

	sc, err := spellchecker.New("abc", spellchecker.WithFilterFunc(fn))
	if err != nil {
		// handle err
	}

	sc.Fix(fn, "word")
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
