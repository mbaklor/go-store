package store

import (
	"sync"
)

type Writable[T any] struct {
	value          T
	nextSubscriber int
	subscribers    map[int]func(T)
	subCh          chan subUpdater[T]
	isRunning      bool
	lock           sync.RWMutex
	wg             sync.WaitGroup
}

func NewWritable[T any](value T) *Writable[T] {
	return &Writable[T]{
		value:       value,
		subscribers: make(map[int]func(T)),
		subCh:       make(chan subUpdater[T]),
		lock:        sync.RWMutex{},
		wg:          sync.WaitGroup{},
	}
}

func (w *Writable[T]) Set(v T) {
	w.lock.Lock()
	w.value = v
	w.lock.Unlock()

	if w.isRunning {
		w.wg.Add(1)
		w.subCh <- newSubCallback[T]()
	}
}

func (w *Writable[T]) Update(updater func(T) T) {
	w.lock.Lock()
	updated := updater(w.value)
	w.lock.Unlock()
	w.Set(updated)
}

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

func (w *Writable[T]) Wait() {
	if w.isRunning {
		w.wg.Wait()
	}
}
