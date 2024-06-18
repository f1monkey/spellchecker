package spellchecker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_priorityQueue(t *testing.T) {
	t.Run("must sort elements by score descending", func(t *testing.T) {
		pq := newPriorityQueue(10)
		pq.Push(match{
			Value: "foo",
			Score: 5,
		})
		pq.Push(match{
			Value: "bar",
			Score: 1,
		})
		pq.Push(match{
			Value: "baz",
			Score: 10,
		})

		require.Equal(t, []match{
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
			pq.Push(match{
				Value: "foo",
				Score: 5,
			})
			pq.Push(match{
				Value: "bar",
				Score: 1,
			})
			pq.Push(match{
				Value: "baz",
				Score: 10,
			})

			require.Equal(t, []match{
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
			pq.Push(match{
				Value: "foo",
				Score: 5,
			})
			pq.Push(match{
				Value: "bar",
				Score: 1,
			})
			pq.Push(match{
				Value: "baz",
				Score: 10,
			})

			require.Equal(t, []match{
				{
					Value: "baz",
					Score: 10,
				},
			}, pq.items)
		})
	})
}
