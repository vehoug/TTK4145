// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
)

func incrementing(ch_incr chan bool, ch_done chan bool) {
	for range 1000000 {
		ch_incr <- true
	}
	ch_done <- true
}

//comment

func decrementing(ch_decr chan bool, ch_done chan bool) {
	for range 1000000 {
		ch_decr <- true
	}
	ch_done <- true
}

func server(ch_incr chan bool, ch_decr chan bool, ch_get chan chan int, ch_quit chan bool) {
	var i = 0
	for {
		select {
		case <-ch_incr:
			i++
		case <-ch_decr:
			i--
		case ch_reply := <-ch_get:
			// Give the server a channel `ch_reply` for it to pass its result through
			ch_reply <- i
		case <-ch_quit:
			return
		}
	}
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	// - GOMAXPROCS decides how many physical CPU cores that can be used in parallel.
	//   If it is set to one (1), the goroutines will create an illusion of parallelism, even though
	//   we are actually just using a single CPU core.
	runtime.GOMAXPROCS(2)

	ch_incr := make(chan bool)
	ch_decr := make(chan bool)
	ch_get := make(chan chan int)
	ch_done := make(chan bool)
	ch_quit := make(chan bool)

	go server(ch_incr, ch_decr, ch_get, ch_quit)

	go incrementing(ch_incr, ch_done)
	go decrementing(ch_decr, ch_done)

	// Wait until both `incrementing` and `decrementing` are finished before continuing
	<-ch_done
	<-ch_done

	// Create a reply channel for the program to receive the final number, and pass
	// the entire channel into the channel `ch_get` which triggers the case in the server.
	ch_reply := make(chan int)
	ch_get <- ch_reply
	final_number := <-ch_reply

	ch_quit <- true

	Println("The magic number is:", final_number)
}
