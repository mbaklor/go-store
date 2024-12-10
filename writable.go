package store

import (
	"github.com/google/uuid"
)

type Writable[T any] struct {
	value       T
	subscribers map[string]func(T)
}

func NewWritable[T any](value T) Writable[T] {
	subscribers := make(map[string]func(T))
	return Writable[T]{value, subscribers}
}

func (w *Writable[T]) Set(v T) {
	w.value = v
	for _, fn := range w.subscribers {
		fn(v)
	}
}

func (w *Writable[T]) Update(updater func(T) T) {
	w.Set(updater(w.value))
}

func (w *Writable[T]) Subscribe(subscriber func(T)) (unsubscriber func()) {
	id := uuid.New().String()
	w.subscribers[id] = subscriber
	return func() {
		delete(w.subscribers, id)
	}
}
