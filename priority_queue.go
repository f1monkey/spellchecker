package spellchecker

import "container/heap"

type priorityQueue struct {
	items    []Match
	capacity int
}

func newPriorityQueue(capacity int) *priorityQueue {
	return &priorityQueue{
		items:    make([]Match, 0, capacity),
		capacity: capacity,
	}
}

func (pq priorityQueue) Len() int { return len(pq.items) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq.items[i].Score < pq.items[j].Score
}

func (pq priorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push unused method. Required Just to implement heap.Interface
func (pq *priorityQueue) Push(x any) {
	item := x.(Match) //nolint:forcetypeassert

	if !pq.accepts(item.Score) {
		return
	}

	pq.insert(item)
}

// Offer adds a word with the given score if it fits into the queue.
// The word is converted to a string only when it is actually stored.
func (pq *priorityQueue) Offer(score float64, word []rune) {
	if !pq.accepts(score) {
		return
	}

	pq.insert(Match{Value: string(word), Score: score})
}

func (pq *priorityQueue) accepts(score float64) bool {
	if pq.capacity <= 0 {
		return false
	}

	return len(pq.items) < pq.capacity || score >= pq.items[0].Score
}

func (pq *priorityQueue) insert(item Match) {
	if len(pq.items) < pq.capacity {
		pq.items = append(pq.items, item)
		heap.Fix(pq, len(pq.items)-1)

		return
	}

	pq.items[0] = item
	heap.Fix(pq, 0)
}

func (pq *priorityQueue) Pop() any {
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[:n-1]

	return item
}

func (pq *priorityQueue) DrainSorted() []Match {
	out := make([]Match, pq.Len())

	for i := len(out) - 1; i >= 0; i-- {
		last := len(pq.items) - 1
		pq.Swap(0, last)
		out[i] = pq.items[last]
		pq.items = pq.items[:last]

		if last > 0 {
			heap.Fix(pq, 0)
		}
	}

	return out
}
