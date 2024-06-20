package pipe

import (
	"container/heap"
)

const (
	defaultShuffleSize = 200
)

type ShuffleStageOption[R any] func(s *shuffleStage[R])

type shuffleStage[R any] struct {
	stopped <-chan struct{}
	queue   *priorityQueue[R]
	bufSize int
}

func ShuffleBuffer[R any](size int) ShuffleStageOption[R] {
	return func(s *shuffleStage[R]) {
		s.bufSize = size
	}
}

func (s *shuffleStage[R]) process(inCh <-chan *R, outCh chan<- *R) {
	defer close(outCh)

	shouldDrain := false
	for {
		if len(s.queue.data) == cap(s.queue.data) || (shouldDrain && s.queue.Len() > 0) {
			peek := s.queue.Peek()
			select {
			case outCh <- peek:
				heap.Pop(s.queue)
			case <-s.stopped:
				return
			}
		} else if s.queue.Len() > 0 {
			select {
			case r, ok := <-inCh:
				if !ok {
					// inCh is closed
					shouldDrain = true
					continue
				}
				heap.Push(s.queue, r)
			default:
				peek := s.queue.Peek()
				select {
				case outCh <- peek:
					heap.Pop(s.queue)
				case newR, ok := <-inCh:
					if !ok {
						// inCh is closed
						shouldDrain = true
						continue
					}
					heap.Push(s.queue, newR)
				case <-s.stopped:
					return
				}
			}
		} else {
			// When queue is empty
			select {
			case newR, ok := <-inCh:
				if !ok {
					// inCh is closed
					return
				}
				heap.Push(s.queue, newR)
			case <-s.stopped:
				return
			}
		}
	}
}

func (s *shuffleStage[R]) getBufSize() int {
	return s.bufSize
}

type priorityQueue[R any] struct {
	data   []*R
	higher func(*R, *R) bool
}

func (pq priorityQueue[R]) Len() int { return len(pq.data) }

func (pq priorityQueue[R]) Less(i, j int) bool {
	return pq.higher(pq.data[i], pq.data[j])
}

func (pq priorityQueue[R]) Swap(i, j int) {
	pq.data[i], pq.data[j] = pq.data[j], pq.data[i]
}

func (pq *priorityQueue[R]) Push(x any) {
	pq.data = append(pq.data, x.(*R))
}

func (pq *priorityQueue[R]) Pop() any {
	n := len(pq.data)
	item := pq.data[n-1]
	pq.data[n-1] = nil // avoid memory leak
	pq.data = pq.data[0 : n-1]
	return item
}

func (pq priorityQueue[R]) Peek() *R {
	return pq.data[0]
}
