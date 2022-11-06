package spellchecker

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_readInput(t *testing.T) {
	t.Run("must use default splitter if it is not provided", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte(`Green. tea!`))

		ch := readInput(buf, nil)

		result := make([]string, 0, 2)
		for item := range ch {
			require.NoError(t, item.err)

			result = append(result, item.word)
		}
		require.Equal(t, []string{"green", "tea"}, result)
	})

	t.Run("must use provided splitter if not nil", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte(`Green tea`))

		ch := readInput(buf, bufio.ScanLines)

		result := make([]string, 0, 2)
		for item := range ch {
			require.NoError(t, item.err)

			result = append(result, item.word)
		}
		require.Equal(t, []string{"Green tea"}, result)
	})
}
