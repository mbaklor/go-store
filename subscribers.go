package store

type subscriberCommand int

const (
	subAdd subscriberCommand = iota
	subRemove
	subCallback
)

type subUpdater[T any] struct {
	command  subscriberCommand
	id       int
	callback func(T)
}

func newSubAdd[T any](id int, fn func(T)) subUpdater[T] {
	return subUpdater[T]{
		command:  subAdd,
		id:       id,
		callback: fn,
	}
}

func newSubRemove[T any](id int) subUpdater[T] {
	return subUpdater[T]{
		command: subRemove,
		id:      id,
	}
}

func newSubCallback[T any]() subUpdater[T] {
	return subUpdater[T]{
		command: subCallback,
	}
}

func (w *Writable[T]) subController() {
	for s := range w.subCh {
		switch s.command {
		case subAdd:
			w.subAdd(s.id, s.callback)
		case subRemove:
			if w.subRemove(s.id) {
				return
			}
		case subCallback:
			w.subCallback()
		}
	}
}

func (w *Writable[T]) subAdd(id int, fn func(T)) {
	w.subscribers[id] = fn
}

func (w *Writable[T]) subRemove(id int) (stop bool) {
	defer w.wg.Done()
	delete(w.subscribers, id)
	if len(w.subscribers) == 0 {
		w.lock.Lock()
		defer w.lock.Unlock()
		w.isRunning = false
		return true
	}
	return false
}

func (w *Writable[T]) subCallback() {
	for _, fn := range w.subscribers {
		w.lock.RLock()
		fn(w.value)
		w.lock.RUnlock()
	}
	w.wg.Done()
}
