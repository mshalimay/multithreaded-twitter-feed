// This script implements a lock-free queue a la Michael and Scott
// Obs: the ABA problem is not addressed in this implementation due to the limitations
// of Go's atomic operation. For more details on that, see https://stackoverflow.com/questions/11525406/atomic-compare-and-swap-with-struct-in-go

package queue

import (
	"sync/atomic"
	"unsafe"
)

// Interface Queue represents a FIFO structure with operations to enqueue and dequeue Requests
type Queue interface {
	Enqueue(*Request)
	Dequeue() *Request
}

// Request represents a client request to be processed by the server
type Request struct {
	Command  	string   	`json:"command"` 	// "ADD", "REMOVE", "CONTAINS", "FEED"
	Id 			int   		`json:"id"`			// unique id for the request
	Body 		string 		`json:"body"`		// the text of the post
	TimeStamp 	float64 	`json:"timestamp"`	// the timestamp of the post
}

// node represents a node in the queue
// Obs: The typical queue would have pointers to other node structs (that is, type next = *node)
// I use unsafe.Pointer to reduce the number of type castings, increasing readability (and potentially speed)
// Repetitive type casting would be necessary if used *node because CAS demands `unsafe.Pointer` arguments.
type node struct{
	request		*Request
	next		unsafe.Pointer 	// pointer to the next node
}


// LockfreeQueue represents a FIFO structure with operations to enqueue
// and dequeue tasks represented as Request
type LockFreeQueue struct {
	head	unsafe.Pointer
	tail 	unsafe.Pointer
}

// NewQueue creates and initializes a LockFreeQueue
func NewLockFreeQueue() Queue {
	// creates a dummy node and pointer to it
	nod := &node{request: nil, next:nil}
	// initial queue: dummy node and head = tail = pointer to dummy node
	return &LockFreeQueue{head:unsafe.Pointer(nod), tail: unsafe.Pointer(nod)}
}


// Enqueue adds a series of Request to the queue
func (queue *LockFreeQueue) Enqueue(task *Request) {
	nod := &node{request:task, next: nil}

	// repeateadly try to enqueue `nod` to the queue and update the tail
	for {
		// get the tail candidate
		tail := (*node)(queue.tail)

		// get currently next node to the tail 
		next := tail.next
		
		// if next is nil, candidate tail potentially the true tail
		if next == nil {
			// compare 'nil' node from candidate tail to current tail again; if succeeds enqueue the new node
			if atomic.CompareAndSwapPointer(&tail.next, next, unsafe.Pointer(nod)){
				return
			}
		// if next is not nil, candidate tail is lagging behind; try to update the tail 
		// (i.e., another thread enqueued successfully but was not able to update the tail poiner; try do the job for him)
		// obs: threads help each other to update the tail 
		// obs: a thread getting stuck means another thread succeeded => the general system made progress
		// obs: in theory, a single thread might get stuck forever updating tails (very unlikely)
		} else {
			// check if q.tail is lagged; if it is, set `next` node as the new tail
			atomic.CompareAndSwapPointer(&queue.tail, unsafe.Pointer(tail), next)
		}
	}
}


// Dequeue removes a Request from the queue
func (queue *LockFreeQueue) Dequeue() *Request {
	for{
		// get head and tail unsafe pointers
		head := queue.head
		tail := queue.tail
		// explanation: unsafe.Pointer is casted to type *node to be able to access the field `next` 
		next := (*node)(head).next

		// head == tail => queue is empty OR tail is lagging behind
		if (head == tail){
			// if next is nil, queue is empty at this snapshot
			if next == nil {
				return nil
			}
			// else, tail is lagging behind; try to update it
			atomic.CompareAndSwapPointer(&queue.tail, tail, next)
		
		// else, try to dequeue 
		} else {	
			// get request data from next node
			request := (*node)(next).request
			// test if head is still head; if not, another thread succeeded in dequeueing => try again
			if atomic.CompareAndSwapPointer(&queue.head, head, next){
				return request
			}
		}
	}
}
