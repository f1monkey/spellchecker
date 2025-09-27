package spellchecker

import (
	"bufio"
	"io"
)

type readData struct {
	word string
	err  error
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
