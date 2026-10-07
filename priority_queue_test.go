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
		pq.Push(Match{
			Value: "foo",
			Score: 5,
		})
		pq.Push(Match{
			Value: "bar",
			Score: 1,
		})
		pq.Push(Match{
			Value: "baz",
			Score: 10,
		})

		require.Equal(t, []Match{
			{
				Value: "bar",
				Score: 1,
			},
			{
				Value: "foo",
				Score: 5,
			},
			{
				Value: "baz",
				Score: 10,
			},
		}, pq.items)
	})

	t.Run("must remove an element with the lowest score if out of capacity", func(t *testing.T) {
		t.Parallel()

		t.Run("2", func(t *testing.T) {
			t.Parallel()

			pq := newPriorityQueue(2)
			pq.Push(Match{
				Value: "foo",
				Score: 5,
			})
			pq.Push(Match{
				Value: "bar",
				Score: 1,
			})
			pq.Push(Match{
				Value: "baz",
				Score: 10,
			})

			require.Equal(t, []Match{
				{
					Value: "foo",
					Score: 5,
				},
				{
					Value: "baz",
					Score: 10,
				},
			}, pq.items)
		})
		t.Run("1", func(t *testing.T) {
			t.Parallel()

			pq := newPriorityQueue(1)
			pq.Push(Match{
				Value: "foo",
				Score: 5,
			})
			pq.Push(Match{
				Value: "bar",
				Score: 1,
			})
			pq.Push(Match{
				Value: "baz",
				Score: 10,
			})

			require.Equal(t, []Match{
				{
					Value: "baz",
					Score: 10,
				},
			}, pq.items)
		})
	})

	t.Run("offer must keep top elements and drain them sorted", func(t *testing.T) {
		t.Parallel()

		pq := newPriorityQueue(2)
		pq.Offer(5, []rune("foo"))
		pq.Offer(1, []rune("bar"))
		pq.Offer(10, []rune("baz"))

		require.Equal(t, []Match{
			{Value: "baz", Score: 10},
			{Value: "foo", Score: 5},
		}, pq.DrainSorted())
	})

	t.Run("offer must ignore elements if capacity is zero", func(t *testing.T) {
		t.Parallel()

		pq := newPriorityQueue(0)
		pq.Offer(5, []rune("foo"))

		require.Empty(t, pq.DrainSorted())
	})
}
