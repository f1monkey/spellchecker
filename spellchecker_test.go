package spellchecker

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadFullSpellchecker() *Spellchecker {
	var s *Spellchecker
	ff, err := os.Open("data/spellchecker.bin")
	if errors.Is(err, os.ErrNotExist) {
		s = newFullSpellchecker()
		dst, err := os.Create("data/spellchecker.bin")
		if err != nil {
			panic(err)
		}

		err = s.Save(dst)
		if err != nil {
			panic(err)
		}
	} else {
		s, err = Load(ff)
		if err != nil {
			panic(err)
		}
	}

	return s
}

func newFullSpellchecker() *Spellchecker {
	f, err := os.Open("data/big.txt")
	if err != nil {
		panic(err)
	}

	s, err := New()
	if err != nil {
		panic(err)
	}

	err = s.AddFrom(f)
	if err != nil {
		panic(err)
	}

	return s
}

func newSampleSpellchecker() *Spellchecker {
	f, err := os.Open("data/sample.txt")
	if err != nil {
		panic(err)
	}

	s, err := New()
	if err != nil {
		panic(err)
	}

	err = s.AddFrom(f)
	if err != nil {
		panic(err)
	}

	return s
}

func Benchmark_Spellchecker_AddFrom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		newFullSpellchecker()
	}
}

func Benchmark_Spellchecker_IsCorrect(b *testing.B) {
	m := loadFullSpellchecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.IsCorrect("tea")
	}
}

func Benchmark_Spellchecker_Fix_3(b *testing.B) {
	m := loadFullSpellchecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Fix("tee")
	}
}

func Benchmark_Spellchecker_Fix_6_Transposition(b *testing.B) {
	m := loadFullSpellchecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Fix("oragne")
	}
}

func Benchmark_Spellchecker_Fix_6_Replacement(b *testing.B) {
	m := loadFullSpellchecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Fix("problam")
	}
}

func Benchmark_Norvig1(b *testing.B) {
	benchmarkNorvig(b, "data/norvig1.txt")
}

func Benchmark_Norvig2(b *testing.B) {
	benchmarkNorvig(b, "data/norvig2.txt")
}

type benchmarkNorvigItem struct {
	expected string
	words    []string
}

func benchmarkNorvig(b *testing.B, dataPath string) {
	b.StopTimer()
	b.ResetTimer()
	m := loadFullSpellchecker()

	testData, err := os.Open(dataPath)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(testData)
	scanner.Split(bufio.ScanLines)

	var data []benchmarkNorvigItem
	for {
		if !scanner.Scan() {
			break
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		line := scanner.Text()

		parts := strings.Split(line, ":")
		required := parts[0]
		checks := strings.Split(parts[1], " ")

		data = append(data, benchmarkNorvigItem{
			expected: required,
			words:    checks,
		})
	}

	total := 0
	ok := 0

	for i := 0; i < b.N; i++ {
		for _, item := range data {
			for _, word := range item.words {
				if word == "" {
					continue
				}

				b.StartTimer()
				result, err := m.Suggest(word, 10)
				b.StopTimer()
				if err != nil && !errors.Is(err, ErrUnknownWord) {
					fmt.Println(err)
				}

				if i == 0 {
					total++
					if len(result) > 0 && result[0] == item.expected {
						ok++
					} else {
						got := ""
						if len(result) > 0 {
							got = result[0]
						}

						fmt.Printf(
							"word %q: expected %q, got %s, all: %v\n",
							word, item.expected, got, result,
						)
					}
				}
			}
		}
	}

	fmt.Printf(
		"Results: %d/%d (%.2f%%)",
		ok, total, float64(ok)/float64(total)*100,
	)
}

func Test_NewSpellchecker(t *testing.T) {
	t.Run("must be able to create a spellchecker without any options", func(t *testing.T) {
		s, err := New()
		require.NoError(t, err)
		require.NotNil(t, s.dict)
	})
	t.Run("must be able to create a spellchecker with custom splitter", func(t *testing.T) {
		s, err := New(WithSplitter(bufio.ScanRunes))
		require.NoError(t, err)
		require.NotNil(t, s.splitter)
	})
}

func Test_Spellchecker_WithOpts(t *testing.T) {
	s, err := New()
	require.NoError(t, err)
	s.WithOpts(WithSplitter(bufio.ScanLines))
	require.NotNil(t, s.splitter)
}

func Test_Spellchecker_IsCorrect(t *testing.T) {
	s := newSampleSpellchecker()

	assert.True(t, s.IsCorrect("orange"))
	assert.False(t, s.IsCorrect("car"))
}

func Test_Spellchecker_Fix(t *testing.T) {
	s := newSampleSpellchecker()
	result, err := s.Fix("problam")
	require.NoError(t, err)
	require.Equal(t, result, "problem")
}
