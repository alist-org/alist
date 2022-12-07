package generic

type Queue[T any] struct {
	queue []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{queue: make([]T, 0)}
}

func (q *Queue[T]) Push(v T) {
	q.queue = append(q.queue, v)
}

func (q *Queue[T]) Pop() T {
	v := q.queue[0]
	q.queue = q.queue[1:]
	return v
}

func (q *Queue[T]) Len() int {
	return len(q.queue)
}

func (q *Queue[T]) IsEmpty() bool {
	return len(q.queue) == 0
}

func (q *Queue[T]) Clear() {
	q.queue = nil
}

func (q *Queue[T]) Peek() T {
	return q.queue[0]
}

func (q *Queue[T]) PeekN(n int) []T {
	return q.queue[:n]
}

func (q *Queue[T]) PopN(n int) []T {
	v := q.queue[:n]
	q.queue = q.queue[n:]
	return v
}

func (q *Queue[T]) PopAll() []T {
	v := q.queue
	q.queue = nil
	return v
}

func (q *Queue[T]) PopWhile(f func(T) bool) []T {
	var i int
	for i = 0; i < len(q.queue); i++ {
		if !f(q.queue[i]) {
			break
		}
	}
	v := q.queue[:i]
	q.queue = q.queue[i:]
	return v
}

func (q *Queue[T]) PopUntil(f func(T) bool) []T {
	var i int
	for i = 0; i < len(q.queue); i++ {
		if f(q.queue[i]) {
			break
		}
	}
	v := q.queue[:i]
	q.queue = q.queue[i:]
	return v
}
