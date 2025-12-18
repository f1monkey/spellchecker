package spellchecker

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Spellchecker_AddFrom(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		sc, err := New("abc")
		require.NoError(t, err)

		buf := bytes.NewBuffer([]byte("hello world"))

		err = sc.AddFrom(buf)
		require.NoError(t, err)

		require.True(t, sc.IsCorrect("world"))

		require.Equal(t, uint(1), sc.dict.counts[1])
		require.Equal(t, uint(1), sc.dict.counts[2])
	})

	t.Run("custom weight", func(t *testing.T) {
		sc, err := New("abc")
		require.NoError(t, err)

		buf := bytes.NewBuffer([]byte("hello world"))

		err = sc.AddFrom(buf, AddWithWeight(2))
		require.NoError(t, err)

		require.True(t, sc.IsCorrect("hello"))
		require.True(t, sc.IsCorrect("world"))

		require.Equal(t, uint(2), sc.dict.counts[1])
		require.Equal(t, uint(2), sc.dict.counts[2])
	})

	t.Run("custom splitter", func(t *testing.T) {
		sc, err := New("abc")
		require.NoError(t, err)

		buf := bytes.NewBuffer([]byte("hello world"))

		err = sc.AddFrom(buf, AddWithSplitter(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) > 0 {
				return len(data), data, nil
			}

			return 0, nil, nil
		}))
		require.NoError(t, err)

		require.False(t, sc.IsCorrect("hello"))
		require.False(t, sc.IsCorrect("world"))
		require.True(t, sc.IsCorrect("hello world"))
	})
}

func Test_Spellchecker_AddMany(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		sc, err := New("abc")
		require.NoError(t, err)

		sc.AddMany([]string{"hello", "world"})

		require.True(t, sc.IsCorrect("hello"))
		require.True(t, sc.IsCorrect("world"))

		require.Equal(t, uint(1), sc.dict.counts[1])
		require.Equal(t, uint(1), sc.dict.counts[2])
	})

	t.Run("custom weight", func(t *testing.T) {
		sc, err := New("abc")
		require.NoError(t, err)

		sc.AddMany([]string{"hello", "world"}, AddWithWeight(2))

		require.True(t, sc.IsCorrect("hello"))
		require.True(t, sc.IsCorrect("world"))

		require.Equal(t, uint(2), sc.dict.counts[1])
		require.Equal(t, uint(2), sc.dict.counts[2])
	})
}

func Test_Spellchecker_Add(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		sc, err := New("abc")
		require.NoError(t, err)

		sc.Add("hello")

		require.True(t, sc.IsCorrect("hello"))

		require.Equal(t, uint(1), sc.dict.counts[1])
	})

	t.Run("custom weight", func(t *testing.T) {
		sc, err := New("abc")
		require.NoError(t, err)

		sc.Add("hello", AddWithWeight(2))

		require.True(t, sc.IsCorrect("hello"))

		require.Equal(t, uint(2), sc.dict.counts[1])
	})
}
