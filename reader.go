package spellchecker

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
)

type readData struct {
	word string
	err  error
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

func readInput(input io.Reader, splitter bufio.SplitFunc) <-chan readData {
	if splitter == nil {
		splitter = defaultSplitter
	}

	ch := make(chan readData)
	scanner := bufio.NewScanner(input)
	scanner.Split(splitter)

	go func() {
		defer close(ch)
		for {
			if !scanner.Scan() {
				break
			}
			if err := scanner.Err(); err != nil {
				ch <- readData{err: err}
				return
			}
			ch <- readData{word: scanner.Text()}
		}
	}()

	return ch
}
