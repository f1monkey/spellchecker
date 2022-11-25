# Spellchecker

Yet another spellchecker written in go.

### Features:
- very small database: approximately 1mb for 30,000 unique words
- time to fix one word - ~35μs

Accuracy in [Peter Norvig's tests](http://norvig.com/spell-correct.html):
* test1 - ~74%
* test2 - ~70%

## Installation

```
$ go get -v github.com/f1monkey/spellchecker
```

## Usage

```go
func main() {
	// Create new instance
	sc, err := spellchecker.New(spellchecker.Alphabet{
		Letters: "abcdefghijklmnopqrstuvwxyz1234567890",
		Length:  36,
	})
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
