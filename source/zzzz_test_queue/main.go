package main

import (
	"fmt"
	"math/rand"
	"proj2/queue"
	"sync"
)

func enqueue(q queue.Queue, req *queue.Request, wg *sync.WaitGroup){
	q.Enqueue(req)
	fmt.Printf("\nEnqueue: %v", req.Id)
	wg.Done()
}

func dequeue(q queue.Queue, wg *sync.WaitGroup){
	req:=q.Dequeue()
	fmt.Printf("\nDequeue: %v", req.Id)
	wg.Done()
}


func main(){

	// create a new queue
	q := queue.NewLockFreeQueue()

	// create a new request
	
	// enqueue the request
	n := 200
	wg := sync.WaitGroup{}
	// wg.Add(n)
	for i:=0; i < n; i++ {
		req := &queue.Request{Command:"test", Id:i, Body:"test", TimeStamp: rand.Float64()}
		// fmt.Printf("\nEnqueue: %v", req.Id)
		q.Enqueue(req)
		// go enqueue(q, req, &wg)
	}
	// wg.Wait()

	wg.Add(n)
	// dequeue the request
	for i:=0; i < n; i++ {
		// req := q.Dequeue()
		// fmt.Printf("\nDequeue: %v", req.Id)
		go dequeue(q, &wg)
	}
	wg.Wait()
	print("\n")
	fmt.Println(q.Dequeue())

	// enqueue again
	for i:=5; i < 10; i++ {
		req := &queue.Request{Command:"test", Id:i, Body:"test", TimeStamp: rand.Float64()}
		q.Enqueue(req)
	}

	// dequeue again
	for i:=5; i < 10; i++ {
		req := q.Dequeue()
		fmt.Printf("\nDequeue: %v", req.Id)
	}
	print("\n")
	fmt.Println(q.Dequeue())

}