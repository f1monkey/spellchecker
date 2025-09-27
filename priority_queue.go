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

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(Match)

	if len(pq.items) < pq.capacity {
		pq.items = append(pq.items, item)
		heap.Fix(pq, len(pq.items)-1)
		return
	}

	if item.Score < pq.items[0].Score {
		return

	}

	pq.items[0] = item
	heap.Fix(pq, 0)
}

func (pq *priorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[:n-1]

	return item
}

func (pq *priorityQueue) DrainSorted() []Match {
	out := make([]Match, pq.Len())

	for i := len(out) - 1; i >= 0; i-- {
		out[i] = heap.Pop(pq).(Match)
	}

	return out
}
