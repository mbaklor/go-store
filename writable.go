package store

import (
	"reflect"
	"sync"
)

type Writable[T any] struct {
	lock        sync.RWMutex
	subID       int
	value       T
	subscribers map[int]func(T)
}

func NewWritable[T any](value T) *Writable[T] {
	subscribers := make(map[int]func(T))
	return &Writable[T]{
		lock:        sync.RWMutex{},
		value:       value,
		subscribers: subscribers,
	}
}

func (w *Writable[T]) Set(v T) {
	w.lock.RLock()
	if eqIgnorePtr(v, w.value) {
		w.lock.RUnlock()
		return
	}
	w.lock.RUnlock()
	w.lock.Lock()
	w.value = v
	w.lock.Unlock()
	w.lock.RLock()
	for _, fn := range w.subscribers {
		fn(v)
	}
	w.lock.RUnlock()
}

func (w *Writable[T]) Update(updater func(T) T) {
	w.lock.Lock()
	newval := updater(w.value)
	w.lock.Unlock()
	w.Set(newval)
}

func (w *Writable[T]) Subscribe(subscriber func(T)) (unsubscriber func()) {
	w.lock.Lock()
	id := w.subID
	w.subID++
	w.subscribers[id] = subscriber
	w.lock.Unlock()
	return func() {
		w.lock.Lock()
		delete(w.subscribers, id)
		w.lock.Unlock()
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
