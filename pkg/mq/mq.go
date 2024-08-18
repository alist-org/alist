package mq

import (
	"sync"

	"github.com/alist-org/alist/v3/pkg/generic"
)

type Message[T any] struct {
	Content T
}

type BasicConsumer[T any] func(Message[T])
type AllConsumer[T any] func([]Message[T])

type MQ[T any] interface {
	Publish(Message[T])
	Consume(BasicConsumer[T])
	ConsumeAll(AllConsumer[T])
	Clear()
	Len() int
}

type inMemoryMQ[T any] struct {
	queue generic.Queue[Message[T]]
	sync.Mutex
}

func NewInMemoryMQ[T any]() MQ[T] {
	return &inMemoryMQ[T]{queue: *generic.NewQueue[Message[T]]()}
}

func (mq *inMemoryMQ[T]) Publish(msg Message[T]) {
	mq.Lock()
	defer mq.Unlock()
	mq.queue.Push(msg)
}

func (mq *inMemoryMQ[T]) Consume(consumer BasicConsumer[T]) {
	mq.Lock()
	defer mq.Unlock()
	for !mq.queue.IsEmpty() {
		consumer(mq.queue.Pop())
	}
}

func (mq *inMemoryMQ[T]) ConsumeAll(consumer AllConsumer[T]) {
	mq.Lock()
	defer mq.Unlock()
	consumer(mq.queue.PopAll())
}

func (mq *inMemoryMQ[T]) Clear() {
	mq.Lock()
	defer mq.Unlock()
	mq.queue.Clear()
}

func (mq *inMemoryMQ[T]) Len() int {
	mq.Lock()
	defer mq.Unlock()
	return mq.queue.Len()
}
