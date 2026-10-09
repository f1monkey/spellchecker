package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_priorityQueue(t *testing.T) {
	t.Parallel()

	t.Run("must sort elements by score descending", func(t *testing.T) {
		t.Parallel()

		pq := newPriorityQueue(10)
		pq.Offer(5, []rune("foo"))
		pq.Offer(1, []rune("bar"))
		pq.Offer(10, []rune("baz"))
		pq.Offer(7, []rune("qux"))
		pq.Offer(3, []rune("quux"))

		require.Equal(t, []Suggestion{
			{Value: "baz", Score: 10},
			{Value: "qux", Score: 7},
			{Value: "foo", Score: 5},
			{Value: "quux", Score: 3},
			{Value: "bar", Score: 1},
		}, pq.DrainSorted())
		require.Equal(t, 0, pq.Len())
	})

	t.Run("must remove an element with the lowest score if out of capacity", func(t *testing.T) {
		t.Parallel()

		t.Run("2", func(t *testing.T) {
			t.Parallel()

			pq := newPriorityQueue(2)
			pq.Offer(5, []rune("foo"))
			pq.Offer(1, []rune("bar"))
			pq.Offer(10, []rune("baz"))

			require.Equal(t, []Suggestion{
				{Value: "baz", Score: 10},
				{Value: "foo", Score: 5},
			}, pq.DrainSorted())
		})

		t.Run("1", func(t *testing.T) {
			t.Parallel()

			pq := newPriorityQueue(1)
			pq.Offer(5, []rune("foo"))
			pq.Offer(1, []rune("bar"))
			pq.Offer(10, []rune("baz"))

			require.Equal(t, []Suggestion{
				{Value: "baz", Score: 10},
			}, pq.DrainSorted())
		})
	})

	t.Run("must replace the minimum with an element of equal score", func(t *testing.T) {
		t.Parallel()

		pq := newPriorityQueue(1)
		pq.Offer(5, []rune("foo"))
		pq.Offer(5, []rune("bar"))

		require.Equal(t, []Suggestion{
			{Value: "bar", Score: 5},
		}, pq.DrainSorted())
	})

	t.Run("must sink through the smaller right child", func(t *testing.T) {
		t.Parallel()

		pq := newPriorityQueue(3)
		pq.Offer(1, []rune("a"))
		pq.Offer(5, []rune("b"))
		pq.Offer(2, []rune("c"))
		pq.Offer(9, []rune("d"))

		require.Equal(t, []Suggestion{
			{Value: "d", Score: 9},
			{Value: "b", Score: 5},
			{Value: "c", Score: 2},
		}, pq.DrainSorted())
	})

	t.Run("must ignore elements if capacity is zero", func(t *testing.T) {
		t.Parallel()

		pq := newPriorityQueue(0)
		pq.Offer(5, []rune("foo"))

		require.Empty(t, pq.DrainSorted())
	})
}
