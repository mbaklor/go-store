package store

import (
	"reflect"
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

type valueCommand int

const (
	valGet valueCommand = iota
	valSet
)

type Writable[T any] struct {
	subID       int
	value       T
	subscribers map[int]func(T)
	subCh       chan subUpdater[T]
	getVal      chan struct{}
	setVal      chan T
	retVal      chan T
}

func (w *Writable[T]) subController() {
	for {
		select {
		case s := <-w.subCh:
			// for s := range w.subCh {
			switch s.command {
			case subAdd:
				w.subscribers[s.id] = s.callback
			case subRemove:
				delete(w.subscribers, s.id)
			case subCallback:
				for _, fn := range w.subscribers {
					fn(s.value)
				}
				// }
			}
		case v := <-w.setVal:
			w.value = v
		case <-w.getVal:
			w.retVal <- w.value
		}
	}
}

func NewWritable[T any](value T) *Writable[T] {
	subscribers := make(map[int]func(T))
	subCh := make(chan subUpdater[T])
	get := make(chan struct{})
	set := make(chan T)
	ret := make(chan T)
	w := &Writable[T]{
		value:       value,
		subscribers: subscribers,
		subCh:       subCh,
		getVal:      get,
		setVal:      set,
		retVal:      ret,
	}
	go w.subController()
	return w
}

func (w *Writable[T]) Set(v T) {
	w.getVal <- struct{}{}
	val := <-w.retVal
	if eqIgnorePtr(v, val) {
		return
	}
	w.setVal <- v

	w.subCh <- subUpdater[T]{
		command: subCallback,
		value:   v,
	}
}

func (w *Writable[T]) Update(updater func(T) T) {
	w.getVal <- struct{}{}
	val := <-w.retVal
	v := updater(val)
	if eqIgnorePtr(v, val) {
		return
	}
	w.setVal <- v

	w.subCh <- subUpdater[T]{
		command: subCallback,
		value:   v,
	}
}

func (w *Writable[T]) Subscribe(subscriber func(T)) (unsubscriber func()) {
	id := w.subID
	w.subID++
	w.subCh <- subUpdater[T]{
		command:  subAdd,
		id:       id,
		callback: subscriber,
	}
	return func() {
		w.subCh <- subUpdater[T]{
			command: subRemove,
			id:      id,
		}
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
