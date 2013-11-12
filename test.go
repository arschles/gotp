package main

func work(recv chan int, wait chan bool, fn func(int)) {
	received := <-recv
	println("work received ", received)
	nextWait := make(chan bool)
	runFn := func(msg int) {
		<-wait
		fn(msg)
		nextWait <- true
	}
	go runFn(received)
	work(recv, nextWait, fn)
}

func main() {
	recv := make(chan int)
	wait := make(chan bool)
	go func() {
		wait <- true
	}()

	worker := func(msg int) {
		println("working on ", msg, "!")
	}
	go work(recv, wait, worker)

	//send stuff forever
	for i := 1; i > 0; i++ {
		recv <- i
	}
}
