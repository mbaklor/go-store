package store

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
