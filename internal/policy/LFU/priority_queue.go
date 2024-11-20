package LFU

import "gocache/internal/policy/interfaces"

type priorityqueue []*lfuEntry

type lfuEntry struct {
	index int
	entry interfaces.Entry
	count int
}

func (l *lfuEntry) Referenced() {
	l.count++
	l.entry.Touch()
}

func (pq priorityqueue) Less(i, j int) bool {
	if pq[i].count == pq[j].count {
		return pq[i].entry.UpdateAt.Before(*pq[j].entry.UpdateAt)
	}
	return pq[i].count < pq[j].count
}

func (pq priorityqueue) Len() int {
	return len(pq)
}

func (pq priorityqueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index, pq[j].index = i, j
}

func (pq *priorityqueue) Pop() interface{} {
	oldpq := *pq
	n := len(oldpq)
	entry := oldpq[n-1]

	//避免内存泄露
	oldpq[n-1] = nil

	newpq := oldpq[:n-1]

	for i := 0; i < len(newpq); i++ {
		newpq[i].index = i
	}
	*pq = newpq
	return entry
}

func (pq *priorityqueue) Push(x interface{}) {
	entry := x.(*lfuEntry)
	entry.index = len(*pq)
	*pq = append(*pq, x.(*lfuEntry))
}
