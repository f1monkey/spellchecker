# Spellchecker

Yet another spellchecker written in go.

- [Spellchecker](#spellchecker)
	- [Features:](#features)
	- [Installation](#installation)
	- [Usage](#usage)
	- [Benchmarks](#benchmarks)
		- [Test set 1:](#test-set-1)
		- [Test set 2:](#test-set-2)

## Features:
- very small database: approximately 1mb for 30,000 unique words
- average time to fix one word  ~35μs
- about 70-74% accuracy in Peter Norvig's test sets (see [benchmarks](#benchmarks))

## Installation

```
$ go get -v github.com/f1monkey/spellchecker
```

## Usage

```go
func main() {
	// Create new instance
	sc, err := spellchecker.New(
		"abcdefghijklmnopqrstuvwxyz1234567890", // allowed symbols, other symbols will be ignored
		spellchecker.WithMaxErrors(2)
	)
	if err != nil {
		panic(err)
	}

	// Read data from any io.Reader
	in, err := os.Open("data/sample.txt")
	if err != nil {
		panic(err)
	}
	sc.AddFrom(in)

	// Add some more words
	sc.Add("lock", "stock", "and", "two", "smoking", "barrels")

	// Check if a word is correct
	result := sc.IsCorrect("coffee")
	fmt.Println(result) // true

	// Fix one word
	fixed, err := sc.Fix("awepon")
	if err != nil && !errors.Is(err, spellchecker.ErrUnknownWord) {
		panic(err)
	}
	fmt.Println(fixed) // weapon

	// Find max=10 suggestions for a word
	matches, err := sc.Suggest("rang", 10)
	if err != nil && !errors.Is(err, spellchecker.ErrUnknownWord) {
		panic(err)
	}
	fmt.Println(matches) // [range, orange]

	// Save data to any io.Writer
	out, err := os.Create("data/out.bin")
	if err != nil {
		panic(err)
	}
	sc.Save(out)

	// Load saved data from io.Reader
	in, err = os.Open("data/out.bin")
	if err != nil {
		panic(err)
	}
	sc, err = spellchecker.Load(in)
	if err != nil {
		panic(err)
	}
}
```

## Benchmarks

Tests are based on data from [Peter Norvig's article about spelling correction](http://norvig.com/spell-correct.html)

#### [Test set 1](http://norvig.com/spell-testset1.txt):

```
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^Benchmark_Norvig1$ github.com/f1monkey/spellchecker

goos: linux
goarch: amd64
pkg: github.com/f1monkey/spellchecker
cpu: 13th Gen Intel(R) Core(TM) i9-13980HX
Benchmark_Norvig1-32    	     294	   3876229 ns/op	        74.07 success_percent	       200.0 success_words	       270.0 total_words	  918275 B/op	    2150 allocs/op
PASS
ok  	github.com/f1monkey/spellchecker	3.378s
```

#### [Test set 2](http://norvig.com/spell-testset2.txt):

```
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^Benchmark_Norvig2$ github.com/f1monkey/spellchecker

goos: linux
goarch: amd64
pkg: github.com/f1monkey/spellchecker
cpu: 13th Gen Intel(R) Core(TM) i9-13980HX
Benchmark_Norvig2-32    	     198	   6102429 ns/op	        70.00 success_percent	       280.0 success_words	       400.0 total_words	 1327385 B/op	    3121 allocs/op
PASS
ok  	github.com/f1monkey/spellchecker	3.895s
```
