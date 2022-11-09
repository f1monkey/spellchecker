package dictionary

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Benchmark_match(b *testing.B) {
	dict := New()
	dict.Add("orange")
	dict.Add("range")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dict.match([]string{"ran", "nge"}, 5)
	}
}

func Test_Match(t *testing.T) {
	dict := New()
	dict.Add("orange")
	dict.Add("ranger")

	t.Run("must return empty bitmap if nothing found", func(t *testing.T) {
		m := dict.match([]string{"qwe"}, 3)
		require.NotNil(t, m)
		require.True(t, m.IsEmpty())
	})
	t.Run("must be able to find word", func(t *testing.T) {
		m := dict.match([]string{"ora", "ran", "ang", "nge"}, 6)
		require.False(t, m.IsEmpty())
		require.Equal(t, uint64(2), m.GetCardinality())
	})
}
