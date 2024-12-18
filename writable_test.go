package store

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWritableInt(t *testing.T) {
	wg := sync.WaitGroup{}
	s := NewWritable(0)
	val := 0
	sum := 0

	unsub := s.Subscribe(func(i int) {
		val = i
		sum += i
		wg.Done()
	})

	wg.Add(1)
	s.Set(2)
	wg.Wait()
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)

	wg.Add(1)
	s.Update(func(i int) int {
		return i * 2
	})
	wg.Wait()
	assert.Equal(t, 4, val)
	assert.Equal(t, 6, sum)

	unsub()
	s.Set(10)
	assert.Equal(t, 4, val)
	assert.Equal(t, 6, sum)
}

func TestWritableStruct(t *testing.T) {
	wg := sync.WaitGroup{}
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
		wg.Done()
	})

	wg.Add(1)
	s.Set(testStruct{2, "first"})
	wg.Wait()
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

	wg.Add(1)
	s.Update(func(ts testStruct) testStruct {
		ts.value = 4
		return ts
	})
	wg.Wait()
	assert.Equal(t, 4, val)
	assert.Equal(t, 6, sum)
	assert.Equal(t, "sub first", word)

	wg.Add(1)
	s.Update(func(ts testStruct) testStruct {
		ts.word = "second"
		return ts
	})
	wg.Wait()
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
	wg := sync.WaitGroup{}
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
		wg.Done()
	})

	wg.Add(1)
	s.Set(&testStruct{2, "first"})
	wg.Wait()
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)
	assert.Equal(t, "sub first", word)

	wg.Add(1)
	s.Update(func(ts *testStruct) *testStruct {
		ts.value = 2
		ts.word = "first"
		return ts
	})
	wg.Wait()
	assert.Equal(t, 2, val)
	assert.Equal(t, 4, sum)
	assert.Equal(t, "sub first", word)

	wg.Add(1)
	s.Update(func(ts *testStruct) *testStruct {
		ts.value = 4
		return ts
	})
	wg.Wait()
	assert.Equal(t, 4, val)
	assert.Equal(t, 8, sum)
	assert.Equal(t, "sub first", word)

	wg.Add(1)
	s.Update(func(ts *testStruct) *testStruct {
		ts.word = "second"
		return ts
	})
	wg.Wait()
	assert.Equal(t, 4, val)
	assert.Equal(t, 12, sum)
	assert.Equal(t, "sub second", word)

	unsub()
	s.Set(&testStruct{10, "third"})
	assert.Equal(t, 4, val)
	assert.Equal(t, 12, sum)
	assert.Equal(t, "sub second", word)
}

func TestWritableAsync(t *testing.T) {
	lock := sync.RWMutex{}
	wg := sync.WaitGroup{}
	type testStruct struct {
		value int
	}
	ts := new(testStruct)
	s := NewWritable(ts)
	val := 0
	unsub := s.Subscribe(func(ts *testStruct) {
		lock.RLock()
		val = ts.value
		lock.RUnlock()
		wg.Done()
	})
	defer unsub()

	wg.Add(100)
	asyncUpdate := func(i int) {
		s.Update(func(ts *testStruct) *testStruct {
			lock.Lock()
			ts.value += i
			lock.Unlock()
			return ts
		})
	}

	for i := 0; i < 100; i++ {
		go asyncUpdate(i)
	}

	wg.Wait()
	assert.Equal(t, 4950, val)
}
