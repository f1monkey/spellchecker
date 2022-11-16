package dictionary

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Benchmark_bitmap_or(b *testing.B) {
	var bm bitmap
	for i := 0; i < b.N; i++ {
		bm.or(1)
	}
}

func Test_bitmap_or(t *testing.T) {
	t.Run("must set positions's bit to 1", func(t *testing.T) {
		var b bitmap
		b.or(0)
		assert.Equal(t, bitmap(1), b)
		b.or(2)
		assert.Equal(t, bitmap(5), b)
	})
	t.Run("must do nothing if bit already is already = 1", func(t *testing.T) {
		var b bitmap
		b.or(0)
		b.or(0)
		assert.Equal(t, bitmap(1), b)
	})
	t.Run("must do nothing on overflow", func(t *testing.T) {
		var b bitmap
		b.or(1000)
		assert.Equal(t, bitmap(0), b)
	})
}

func Test_bitmap_clone(t *testing.T) {
	t.Run("must return new bitmap's instance", func(t *testing.T) {
		var b1 bitmap
		b1.or(0)

		b2 := b1.clone()
		b2.or(2)
		assert.Equal(t, bitmap(5), b2)
		assert.Equal(t, bitmap(1), b1)
	})
}
