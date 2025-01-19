package store

import (
	"sync"
)

type subscriberCommand int

const (
	subAdd subscriberCommand = iota
	subRemove
	subCallback
)

type subUpdater[T any] struct {
	id       int
	command  subscriberCommand
	callback func(T)
	value    T
}

type Writable[T any] struct {
	lock        sync.RWMutex
	subID       int
	value       T
	subscribers map[int]func(T)
	subCh       chan subUpdater[T]
	wg          sync.WaitGroup
	running     bool
}

func (w *Writable[T]) subController() {
	for s := range w.subCh {
		switch s.command {
		case subAdd:
			w.subscribers[s.id] = s.callback
		case subRemove:
			delete(w.subscribers, s.id)
			if len(w.subscribers) == 0 {
				w.running = false
				w.wg.Done()
				return
			}
			w.wg.Done()
		case subCallback:
			for _, fn := range w.subscribers {
				fn(s.value)
			}
			w.wg.Done()
		}
	}
}

func NewWritable[T any](value T) *Writable[T] {
	subscribers := make(map[int]func(T))
	subCh := make(chan subUpdater[T])
	w := &Writable[T]{
		lock:        sync.RWMutex{},
		value:       value,
		subscribers: subscribers,
		subCh:       subCh,
		wg:          sync.WaitGroup{},
	}
	return w
}

func (w *Writable[T]) Set(v T) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.value = v

	if w.running {
		w.wg.Add(1)
		w.subCh <- subUpdater[T]{
			command: subCallback,
			value:   v,
		}
	}
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
	w.lock.Unlock()
	if !w.running {
		go w.subController()
		w.running = true
	}
	w.subCh <- subUpdater[T]{
		command:  subAdd,
		id:       id,
		callback: subscriber,
	}
	return func() {
		w.wg.Add(1)
		w.subCh <- subUpdater[T]{
			command: subRemove,
			id:      id,
		}
		w.wg.Wait()
	}
}

func (w *Writable[T]) Wait() {
	if w.running {
		w.wg.Wait()
	}
}
