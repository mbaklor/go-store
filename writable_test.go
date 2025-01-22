package store_test

import (
	"sync"
	"testing"

	"github.com/mbaklor/go-store"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptions(t *testing.T) {
	s := store.NewWritable(0)
	sum := 0

	unsub := s.Subscribe(func(i int) {
		sum += i
	})
	assert.Equal(t, 0, sum)

	s.Set(2)
	s.Wait()
	assert.Equal(t, 2, sum)

	unsub()
	s.Set(2)
	s.Wait()
	assert.Equal(t, 2, sum)

	unsub = s.Subscribe(func(i int) {
		sum += i
	})
	assert.Equal(t, 4, sum)

	s.Set(2)
	s.Wait()
	assert.Equal(t, 6, sum)

	unsub()
	s.Set(2)
	s.Wait()
	assert.Equal(t, 6, sum)

	unsub1 := s.Subscribe(func(i int) {
		sum += i
	})
	assert.Equal(t, 8, sum)
	unsub2 := s.Subscribe(func(i int) {
		sum += i
	})
	assert.Equal(t, 10, sum)

	s.Set(2)
	s.Wait()
	assert.Equal(t, 14, sum)

	unsub1()
	s.Set(2)
	s.Wait()
	assert.Equal(t, 16, sum)

	unsub2()
	s.Set(2)
	s.Wait()
	assert.Equal(t, 16, sum)
}

func TestWritableInt(t *testing.T) {
	s := store.NewWritable(0)
	val := 0
	sum := 0

	unsub := s.Subscribe(func(i int) {
		val = i
		sum += i
	})

	s.Set(2)
	s.Wait()
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)

	s.Update(func(i int) int {
		return i * 2
	})
	s.Wait()
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
	s := store.NewWritable(testStruct{0, "test"})
	val := 0
	sum := 0
	word := ""

	unsub := s.Subscribe(func(ts testStruct) {
		val = ts.value
		sum += ts.value
		word = "sub " + ts.word
	})

	s.Set(testStruct{2, "first"})
	s.Wait()
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts testStruct) testStruct {
		ts.value = 2
		ts.word = "first"
		return ts
	})
	s.Wait()
	assert.Equal(t, 2, val)
	assert.Equal(t, 4, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts testStruct) testStruct {
		ts.value = 4
		return ts
	})
	s.Wait()
	assert.Equal(t, 4, val)
	assert.Equal(t, 8, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts testStruct) testStruct {
		ts.word = "second"
		return ts
	})
	s.Wait()
	assert.Equal(t, 4, val)
	assert.Equal(t, 12, sum)
	assert.Equal(t, "sub second", word)

	unsub()
	s.Set(testStruct{10, "third"})
	assert.Equal(t, 4, val)
	assert.Equal(t, 12, sum)
	assert.Equal(t, "sub second", word)
}

func TestWritablePointer(t *testing.T) {
	type testStruct struct {
		value int
		word  string
	}
	s := store.NewWritable(new(testStruct))
	val := 0
	sum := 0
	word := ""

	unsub := s.Subscribe(func(ts *testStruct) {
		val = ts.value
		sum += ts.value
		word = "sub " + ts.word
	})

	s.Set(&testStruct{2, "first"})
	s.Wait()
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts *testStruct) *testStruct {
		ts.value = 2
		ts.word = "first"
		return ts
	})
	s.Wait()
	assert.Equal(t, 2, val)
	assert.Equal(t, 4, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts *testStruct) *testStruct {
		ts.value = 4
		return ts
	})
	s.Wait()
	assert.Equal(t, 4, val)
	assert.Equal(t, 8, sum)
	assert.Equal(t, "sub first", word)

	s.Update(func(ts *testStruct) *testStruct {
		ts.word = "second"
		return ts
	})
	s.Wait()
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
	s := store.NewWritable(ts)
	val := 0
	wg.Add(1)
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
