package spellchecker

import "container/heap"

// priorityQueue implements heap.Interface and holds matches.
type priorityQueue struct {
	items    []match
	capacity int
}

// newPriorityQueue initializes a new priorityQueue with a given capacity
func newPriorityQueue(capacity int) *priorityQueue {
	return &priorityQueue{
		items:    make([]match, 0, capacity),
		capacity: capacity,
	}
}

func (pq priorityQueue) Len() int { return len(pq.items) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq.items[j].Score > pq.items[i].Score
}

func (pq priorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(match)
	if len(pq.items) < pq.capacity {
		pq.items = append(pq.items, item)
		heap.Fix(pq, len(pq.items)-1)
	} else if len(pq.items) > 0 && item.Score >= pq.items[0].Score {
		pq.items[0] = item
		heap.Fix(pq, 0)
	}
}

func (pq *priorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[0 : n-1]
	return item
}
