package spellchecker

// priorityQueue keeps the top items by score.
type priorityQueue struct {
	items    []Suggestion
	capacity int
}

func newPriorityQueue(capacity int) *priorityQueue {
	return &priorityQueue{
		items:    make([]Suggestion, 0, capacity),
		capacity: capacity,
	}
}

func (pq *priorityQueue) Len() int { return len(pq.items) }

// Offer adds the word if its score is high enough.
func (pq *priorityQueue) Offer(score float64, word []rune) {
	if pq.capacity <= 0 {
		return
	}

	if len(pq.items) < pq.capacity {
		pq.items = append(pq.items, Suggestion{Value: string(word), Score: score})
		pq.up(len(pq.items) - 1)

		return
	}

	if score < pq.items[0].Score {
		return
	}

	pq.items[0] = Suggestion{Value: string(word), Score: score}
	pq.down(0)
}

// DrainSorted empties the queue and returns its items, best first.
func (pq *priorityQueue) DrainSorted() []Suggestion {
	out := make([]Suggestion, len(pq.items))

	for i := len(out) - 1; i >= 0; i-- {
		last := len(pq.items) - 1
		pq.items[0], pq.items[last] = pq.items[last], pq.items[0]
		out[i] = pq.items[last]
		pq.items = pq.items[:last]
		pq.down(0)
	}

	return out
}

func (pq *priorityQueue) up(i int) {
	for i > 0 {
		parent := (i - 1) >> 1
		if pq.items[i].Score >= pq.items[parent].Score {
			return
		}

		pq.items[i], pq.items[parent] = pq.items[parent], pq.items[i]

		i = parent
	}
}

func (pq *priorityQueue) down(i int) {
	n := len(pq.items)

	for {
		smallest := i<<1 + 1
		if smallest >= n {
			return
		}

		if right := smallest + 1; right < n && pq.items[right].Score < pq.items[smallest].Score {
			smallest = right
		}

		if pq.items[smallest].Score >= pq.items[i].Score {
			return
		}

		pq.items[i], pq.items[smallest] = pq.items[smallest], pq.items[i]

		i = smallest
	}
}
