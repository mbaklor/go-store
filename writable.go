package store

import (
	"sync"
)

// Writable is a go implementation of svelte writable stores
//
// https://svelte.dev/docs/svelte/stores#svelte-store-writable
type Writable[T any] struct {
	value          T
	nextSubscriber int
	subscribers    map[int]func(T)
	subCh          chan subUpdater[T]
	isRunning      bool
	lock           sync.RWMutex
	wg             sync.WaitGroup
}

// NewWritable creates a new store of type T instantiated with an inital value
func NewWritable[T any](value T) *Writable[T] {
	return &Writable[T]{
		value:       value,
		subscribers: make(map[int]func(T)),
		subCh:       make(chan subUpdater[T]),
		lock:        sync.RWMutex{},
		wg:          sync.WaitGroup{},
	}
}

// Set takes a value v, overrides the store's value and calls all subscribers
func (w *Writable[T]) Set(v T) {
	w.lock.Lock()
	w.value = v
	w.lock.Unlock()

	if w.isRunning {
		w.wg.Add(1)
		w.subCh <- newSubCallback[T]()
	}
}

// Update updates the store's value by running it through a supplied updater
// func and calls all subscribers
func (w *Writable[T]) Update(updater func(T) T) {
	w.lock.Lock()
	updated := updater(w.value)
	w.lock.Unlock()
	w.Set(updated)
}

// Subscribe takes a func that is fired every time the internal value is
// changed by a Set or Update. It returns an unsubscribe func to
// clean up the subscriber when done using it.
func (w *Writable[T]) Subscribe(subscriber func(T)) (unsubscriber func()) {
	w.lock.Lock()
	id := w.nextSubscriber
	w.nextSubscriber++
	if !w.isRunning {
		go w.subController()
		w.isRunning = true
	}
	w.lock.Unlock()
	w.subCh <- newSubAdd(id, subscriber)

	w.lock.RLock()
	subscriber(w.value)
	w.lock.RUnlock()

	return func() {
		w.wg.Add(1)
		w.subCh <- newSubRemove[T](id)
		w.wg.Wait()
	}
}

// Wait is used to wait for subscribers to finish running if syncronous
// work is needed.
func (w *Writable[T]) Wait() {
	if w.isRunning {
		w.wg.Wait()
	}
}
