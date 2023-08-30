// This script implements a simple sequential queue using a linked list
// serves as a basis to the lock-free queue

package queue

type qNode struct {
	request 	*Request
	next 		*qNode
}

type SeqQueue struct {
	head 	*qNode
	tail 	*qNode
}

// NewQueue creates and initializes a LockFreeQueue
func NewSeqQueue() Queue {
	// create a dummy node and get a pointer to it
	QNode := &qNode{nil, nil}
	// initial queue = dummy node and head = tail = pointer to dummy node
	return &SeqQueue{head:QNode, tail: QNode}
}

// Enqueue adds Request to the queue
func (q *SeqQueue) Enqueue(request *Request) {
	qnode := &qNode{request: request, next:nil}
	q.tail.next = qnode
	q.tail = qnode
}

// Dequeue a Request from the queue
func (q *SeqQueue) Dequeue() *Request {
	if q.head == q.tail {
		if q.head.next == nil {
			return nil
		}
	}
	request := q.head.next.request
	q.head = q.head.next
	return request
}