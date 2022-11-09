package dictionary

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Benchmark_Match(b *testing.B) {
	dict := New()
	dict.Add("orange")
	dict.Add("range")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dict.Match("range")
	}
}

func Test_Match(t *testing.T) {
	dict := New()
	dict.Add("orange")
	dict.Add("ranger")

	t.Run("must return empty bitmap if nothing found", func(t *testing.T) {
		m, err := dict.Match("qwe")
		require.NoError(t, err)
		require.NotNil(t, m)
		require.True(t, m.IsEmpty())
	})
	t.Run("must be able to find word", func(t *testing.T) {
		m, err := dict.Match("orange")
		require.NoError(t, err)
		require.False(t, m.IsEmpty())
		require.Equal(t, uint64(2), m.GetCardinality())
	})
}
