package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_priorityQueue(t *testing.T) {
	t.Run("must sort elements by score descending", func(t *testing.T) {
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
		t.Run("2", func(t *testing.T) {
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
}
