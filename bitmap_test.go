package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Benchmark_bitmap_or(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var bm bitmap
		bm.or(1)
	}
}

func Test_bitmap_or(t *testing.T) {
	t.Run("must set specified bit to 1", func(t *testing.T) {
		var b bitmap
		b.or(0)
		assert.Equal(t, bitmap(1), b)
		b.or(2)
		assert.Equal(t, bitmap(5), b)
	})
	t.Run("must do nothing if the specified bit is already == 1", func(t *testing.T) {
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

func Benchmark_bitmap_xor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var bm bitmap
		bm.xor(1)
	}
}

func Test_bitmap_xor(t *testing.T) {
	t.Run("must invert bit", func(t *testing.T) {
		var b bitmap
		b.xor(0)
		assert.Equal(t, bitmap(1), b)
		b.xor(0)
		assert.Equal(t, bitmap(0), b)
	})
	t.Run("must do nothing on overflow", func(t *testing.T) {
		var b bitmap
		b.xor(1000)
		assert.Equal(t, bitmap(0), b)
	})
}

func Test_bitmap_isEmpty(t *testing.T) {
	t.Run("must return true if it's empty", func(t *testing.T) {
		var b bitmap
		require.True(t, b.isEmpty())
	})
	t.Run("must return false if it is not empty", func(t *testing.T) {
		var b bitmap
		b.or(1)
		require.False(t, b.isEmpty())
	})
}

func Benchmark_bitmap_countDiff(b *testing.B) {
	var b1, b2 bitmap
	b1.or(1)
	b1.or(2)
	b2.or(31)
	for i := 0; i < b.N; i++ {
		b1.countDiff(b2)
	}
}

func Test_bitmap_countDiff(t *testing.T) {
	t.Run("must return 0 if bitmaps are equal", func(t *testing.T) {
		var b1, b2 bitmap
		b1.or(1)
		b2.or(1)
		require.Equal(t, 0, b1.countDiff(b2))
	})

	t.Run("must return correct count of different bits", func(t *testing.T) {
		var b1, b2 bitmap
		b1.or(1)
		b1.or(2)
		b1Check := b1.clone()
		b2.or(31)
		b2Check := b2.clone()
		require.Equal(t, 3, b1.countDiff(b2))
		require.Equal(t, b1, b1Check)
		require.Equal(t, b2, b2Check)
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
