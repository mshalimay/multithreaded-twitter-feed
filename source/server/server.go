package server

import (
	"encoding/json"
	"fmt"
	"sync"
	"proj2/feed"
	"proj2/queue"
	"proj2/lock"
)

// Represents a response to a client request for "ADD", "REMOVE", "CONTAINS", "FEED"
type Response struct {
	Success bool `json:"success"`
	Id      int  `json:"id"`
}

// Represents a response to a client request for "FEED"
type FeedResponse struct {
	Id      int 		`json:"id"`
	Feed 	[]feed.Post `json:"feed"` 
			// feed.Post contains the `body` and `timestamp` of a post; see feed/feed.go	
}

type Config struct {
	Encoder *json.Encoder // Represents the buffer to encode Responses
	Decoder *json.Decoder // Represents the buffer to decode Requests
	Mode    string        // Represents whether the server should execute
	// sequentially or in parallel
	// If Mode == "s"  then run the sequential version
	// If Mode == "p"  then run the parallel version
	// These are the only values for Version
	ConsumersCount int // Represents the number of consumers to spawn
}


// SyncContext is a struct that contains synchronization constructs for the consumer-producer model
type SyncContext struct {
	// mux 			sync.Mutex			// A mutual exclusion lock
	mux 			lock.DummyLocker	// A Locker whose Lock() and Unlock() methods do nothing
	cond 			*sync.Cond			// cond is used for producer to signal consumer of tasks and consumer to wait for tasks
	wg 				sync.WaitGroup		// wg keeps track of the number of tasks remaining to be executed
}
// Obs: dummylock is used because we only need the signaling and enqueing aspects of the condition variable.
// One can also use the traditional mutex, but would have the overhead of Lock() and Unlock() unnecessarily.


// NewContext creates and initializes a SyncContext
func NewContext() *SyncContext {
	ctx := &SyncContext{}
	ctx.cond = sync.NewCond(&ctx.mux)
	return ctx
}

//Run starts up the twitter server based on the configuration information
// provided and only returns when the server is fully shutdown.
func Run(config Config) {
	// create a new feed
	f := feed.NewFeed() 		// naive coarse-grained locking
	// f := feed.NewOptFeed()	// optimistic locking
	
	// run the server in sequential mode
	if config.Mode == "s" {
		RunSequential(f, config.Encoder, config.Decoder)
	
	// run the server in parallel mode
	} else {
		// create a new lock-free queue and sync context
		q := queue.NewLockFreeQueue()
		ctx := NewContext()	
		// spawn the consumers as separate goroutines
		for i:=0; i < config.ConsumersCount; i++{
			go consumer(f, config.Encoder, q, ctx)
		}
		// start the producer
		producer(f, config.Decoder, config.Encoder, q, ctx)
	}
}


func producer(f feed.Feed, dec *json.Decoder, enc *json.Encoder, q queue.Queue, ctx *SyncContext) {
	
	// loops reading requests from os.Stdin until the client sends a "DONE" request 
	for {
		// decode the request
		request := &queue.Request{}
 		err := dec.Decode(&request)
		
		if err != nil {
			fmt.Printf("\nError decoding request: %s\n", err.Error())
			return
		}
		// if "DONE" command, wait for consumers to finish remaining tasks and shutdown the server
		if request.Command == "DONE" {
			ctx.wg.Wait()
			return
		}
		// add a task to the wg and enqueue it
		ctx.wg.Add(1)
		q.Enqueue(request)

		// signal the consumers that there is a new task
		ctx.cond.Broadcast()
	}
}

// consumer waits for tasks to be enqueued and executes them.
func consumer(f feed.Feed, enc *json.Encoder, q queue.Queue, ctx *SyncContext) {	
	for {
		// try to dequeue a task
		task := q.Dequeue()		
		
		// if the queue is empty, wait for the producer to enqueue a task
		if task == nil {
			ctx.mux.Lock()
			ctx.cond.Wait()
			ctx.mux.Unlock()		
		// if task retrieved, execute it, subtract from the wg and try to dequeue another task
		} else {
			execute(f, enc, task)
			ctx.wg.Done()
		}
	}
}

// execute executes a task = client request and sends the response to the client
func execute(f feed.Feed, enc *json.Encoder, task *queue.Request) {
	switch task.Command{
	case "ADD":	
		f.Add(task.Body, task.TimeStamp)
		enc.Encode(Response{Success: true, Id: task.Id})

	case "REMOVE":
		success := f.Remove(task.TimeStamp)
		enc.Encode(Response{Success: success, Id: task.Id})

	case "CONTAINS":
		success := f.Contains(task.TimeStamp)
		enc.Encode(Response{Success: success, Id: task.Id})

	case "FEED":
		feedPosts := f.ReturnFeed()
		enc.Encode(FeedResponse{Id: task.Id, Feed: feedPosts})
		return
	}
}

// RunSequential runs the server in sequential mode
func RunSequential(f feed.Feed, enc *json.Encoder, dec *json.Decoder) {
	var request queue.Request
	for {
		// decode the request
 		err := dec.Decode(&request)

		if err != nil {
			fmt.Printf("\nError decoding request: %s\n", err.Error())
			return
		}

		// if "DONE" command, shutdown the server
		if request.Command == "DONE" {
			return
		}

		// execute the request
		execute(f, enc, &request)
	}
}

