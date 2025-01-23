# Go Stores

go-stores is a port of svelte's [reactive stores](https://svelte.dev/docs/svelte/stores) which allow you to create a data store which can be set and updated, while notifying listeners of the change, allowing for asynchronous data binding and UI updates.


## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/mbaklor/go-stores"
)

func main() {
	val := 0
	s := store.NewWritable(0)

	// Create a subscriber that will update our variable val with the
	// updated value in the store, and print to console that there was
	// an update
	// subscribers fire the first time they're added
	unsub := s.Subscribe(func(i int) {
		val = i
		fmt.Printf("Store updated, new value %d\n", val)
	})
	defer unsub() // don't forget to clean up!

	// val now is also 2
	s.Set(2)

	// if you want the stores updating synchronously, just wait for them!
	s.Wait()

	// I don't know what the value of the store is, but please double it!
	// in this case it was 2 so val in now 4
	s.Update(func(i int) int {
		return i * 2
	})

	// some subscribers might need to run long lasting tasks!
	// let's emulate that with a sleep
	unsub2 := s.Subscribe(func(i int) {
		time.Sleep(time.Second * 3)
		fmt.Printf("long lasting proccess on value %d\n", i)
	})
	defer unsub2() // yup, this one too

	s.Set(3)

	// subscribers update unsorted, so we won't know if val changed
	// before or after the long process, so once again, we can just wait!
	s.Wait()

	fmt.Printf("End of the program, val = %d\n", val)

	// Output:
	// Store updated, new value 0
	// Store updated, new value 2
	// Store updated, new value 4
	// long lasting proccess on value 4
	// Store updated, new value 3
	// long lasting proccess on value 3
	// End of the program, val = 3
}
```
