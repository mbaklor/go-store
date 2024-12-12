package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWritableInt(t *testing.T) {
	s := NewWritable(0)
	val := 0
	sum := 0

	unsub := s.Subscribe(func(i int) {
		val = i
		sum += i
	})

	s.Set(2)
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)

	s.Update(func(i int) int {
		return i * 2
	})
	assert.Equal(t, 4, val)
	assert.Equal(t, 6, sum)

	unsub()
	s.Set(10)
	assert.Equal(t, 4, val)
	assert.Equal(t, 6, sum)
}

func TestWritableStruct(t *testing.T) {
	type testStruct struct {
		value int
		word  string
	}
	s := NewWritable(testStruct{0, "test"})
	val := 0
	sum := 0
	word := ""

	unsub := s.Subscribe(func(ts testStruct) {
		val = ts.value
		sum += ts.value
		word = "sub " + ts.word
	})

	s.Set(testStruct{2, "first"})
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts testStruct) testStruct {
		ts.value = 2
		ts.word = "first"
		return ts
	})
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts testStruct) testStruct {
		ts.value = 4
		return ts
	})
	assert.Equal(t, 4, val)
	assert.Equal(t, 6, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts testStruct) testStruct {
		ts.word = "second"
		return ts
	})
	assert.Equal(t, 4, val)
	assert.Equal(t, 10, sum)
	assert.Equal(t, "sub second", word)

	unsub()
	s.Set(testStruct{10, "third"})
	assert.Equal(t, 4, val)
	assert.Equal(t, 10, sum)
	assert.Equal(t, "sub second", word)
}

func TestWritablePointer(t *testing.T) {
	type testStruct struct {
		value int
		word  string
	}
	s := NewWritable(new(testStruct))
	val := 0
	sum := 0
	word := ""

	unsub := s.Subscribe(func(ts *testStruct) {
		val = ts.value
		sum += ts.value
		word = "sub " + ts.word
	})

	s.Set(&testStruct{2, "first"})
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts *testStruct) *testStruct {
		ts.value = 2
		ts.word = "first"
		return ts
	})
	assert.Equal(t, 2, val)
	assert.Equal(t, 4, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts *testStruct) *testStruct {
		ts.value = 4
		return ts
	})
	assert.Equal(t, 4, val)
	assert.Equal(t, 8, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts *testStruct) *testStruct {
		ts.word = "second"
		return ts
	})
	assert.Equal(t, 4, val)
	assert.Equal(t, 12, sum)
	assert.Equal(t, "sub second", word)

	unsub()
	s.Set(&testStruct{10, "third"})
	assert.Equal(t, 4, val)
	assert.Equal(t, 12, sum)
	assert.Equal(t, "sub second", word)
}
