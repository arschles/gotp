package main

import "fmt"

func work(order int, c chan int) {
	fmt.Printf("starting worker %d\n", order)
	recvOrder := <-c
	for recvOrder != order {
		fmt.Printf("worker %d still waiting for its order", order)
		recvOrder := <-c
	}
	fmt.Printf("worker %d completed, sending %d to channel\n", order, order+1)
	c <- (order + 1)
}

func main() {
	c := make(chan int)

	go work(0, c)
	go work(1, c)
	go work(2, c)
	c <- 0
	<-c
}
