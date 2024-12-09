package store

type Writable[T any] struct {
	value       T
	subscribers map[string]func(T)
}

func NewWritable[T any](value T) Writable[T] {
	subscribers := make(map[string]func(T))
	return Writable[T]{value, subscribers}
}
