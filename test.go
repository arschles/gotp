package main

import "fmt"

func work(order int, waiter chan int, next chan int) {
	fmt.Printf("starting worker %d\n", order)
	<-(waiter)
	fmt.Printf("worker %d completed\n", order)
	(next) <- 1
	fmt.Printf("worker %d finished sending to next chan\n", order)
}

func main() {
	c1 := make(chan int)
	c2 := make(chan int)
	c3 := make(chan int)
	c4 := make(chan int)
	go work(0, c1, c2)
	go work(1, c2, c3)
	go work(2, c3, c4)
	c1 <- 0
	<-c4
}
