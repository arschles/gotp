package main

import "fmt"

func work(num int, recv chan int, next chan int, done chan int, run func()) {
	fmt.Printf("created worker %d\n", num)
	<-recv
	fmt.Printf("starting worker %d\n", num)
	run()
	done <- 1
	next <- 1
	fmt.Printf("finished worker %d\n", num)
}

var recv chan int
var next chan int
var workerNum int

//schedule a function to run.
//returns the channel to which the worker will send when it's done
func schedule(f func()) chan int {
	done := make(chan int)
	go work(workerNum, recv, next, done, f)
	if workerNum == 0 {
		recv <- 1
	}
	workerNum++
	recv = make(chan int)
	next = make(chan int)
	return done
}

func main() {
	recv = make(chan int)
	next = make(chan int)
	workerNum = 1

	c1 := schedule(func() {
		fmt.Println("doing stuff1")
	})

	c2 := schedule(func() {
		fmt.Println("doing stuff2")
	})

	c3 := schedule(func() {
		fmt.Println("doing stuff3")
	})

	c4 := schedule(func() {
		fmt.Println("doing stuff4")
	})

	<-c1
	<-c2
	<-c3
	<-c4
}
