package main

import (
	"fmt"
	"sync"
)

// Test channel

func main() {
	var Pending chan int = make(chan int)
	var Quit chan int = make(chan int)
	var wg sync.WaitGroup

	wg.Add(1)
	go func(p chan int, q chan int) {

		for {
			//time.Sleep(100 * time.Millisecond)
			select {
			case number := <-p:
				fmt.Printf("Number A %d \n", number)
				if number == 10000 {
					q <- 1
					return
				}

			}
		}
	}(Pending, Quit)

	for i := 0; i < 100; i++ {

		Pending <- i

	}
	<-Quit
	//Quit <- 1

}
