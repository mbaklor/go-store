package store

import (
	"reflect"
)

type Writable[T any] struct {
	subCount    int
	value       T
	subscribers map[int]func(T)
}

func NewWritable[T any](value T) Writable[T] {
	subscribers := make(map[int]func(T))
	return Writable[T]{0, value, subscribers}
}

func (w *Writable[T]) Set(v T) {
	if eqIgnorePtr(v, w.value) {
		return
	}
	w.value = v
	for _, fn := range w.subscribers {
		fn(v)
	}
}

func (w *Writable[T]) Update(updater func(T) T) {
	w.Set(updater(w.value))
}

func (w *Writable[T]) Subscribe(subscriber func(T)) (unsubscriber func()) {
	id := w.subCount
	w.subCount++
	w.subscribers[id] = subscriber
	return func() {
		delete(w.subscribers, id)
	}
}

// Check equality of a and b
//
// always return true if T is a pointer
// because Update in a pointer struct will mutate the original struct
func eqIgnorePtr[T any](a, b T) bool {
	if reflect.ValueOf(a).Kind() == reflect.Pointer {
		return false
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	return false
}
