package queue

// Tests for correctness of the lockfree queue implementation

import (
	"fmt"
	"sync"
	"testing"
)



func isPermutation(n int, data []int) bool {
	seen := make([]bool, n)

	for i := 0; i < n; i++ {
		value := data[i]

		if value < 0 || value >= n {
			return false
		}
		if seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func TestSequential(t *testing.T) {
	q := NewLockFreeQueue()

	for i := 0; i < 10; i++ {
		q.Enqueue(&Request{Id: i})
	}
	for i := 0; i < 10; i++ {
		result := q.Dequeue()

		if result == nil || result.Id != i {
			t.Errorf("Expected %d, got %v", i, result)
		}
	}
}

func TestPermute(t *testing.T) {
	n := 100
	q := NewLockFreeQueue()
	data := make([]int, n)

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			q.Enqueue(&Request{Id: i})
			wg.Done()
		}(i)
	}

	wg.Wait()

	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			result := q.Dequeue()
			if result != nil {
				data[i] = result.Id
				
			} else {
				fmt.Println("\nNil value!")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
	fmt.Printf("\nData: %v, len %d", data, len(data))

	if !isPermutation(n, data) {
		t.Errorf("Expected permutation, got %v", data)
	}
}

func assertFIFO(nthreads, nvals int, data []int) bool {
	expect := make([]int, nthreads)

	for tid := 0; tid < nthreads; tid++ {
		expect[tid] = tid * nvals
	}
	for i := 0; i < nthreads*nvals; i++ {
		x := data[i]
		tid := x / nvals

		if expect[tid] != x {
			return false
		}
		expect[tid]++
	}
	return true
}

func TestFIFO(t *testing.T) {
	nthreads := 10
	nvals := 200
	result := make([]int, nthreads*nvals)

	q := NewLockFreeQueue()

	var wg sync.WaitGroup
	wg.Add(nthreads)

	for tid := 0; tid < nthreads; tid++ {
		go func(tid int) {
			for i := 0; i < nvals; i++ {
				q.Enqueue(&Request{Id: i + tid*nvals})
			}
			wg.Done()
		}(tid)
	}

	wg.Wait()

	for i := 0; i < nthreads*nvals; i++ {
		r := q.Dequeue()
		if r != nil {
			result[i] = r.Id
		}
	}

	if !assertFIFO(nthreads, nvals, result) {
		t.Errorf("Expected FIFO, got %v", result)
	}
}


func TestMultiGoroutineEnqueue(t *testing.T) {
	q := NewLockFreeQueue()
	wg := &sync.WaitGroup{}
	n := 1000
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < n; j++ {
				req := &Request{Id: j}
				q.Enqueue(req)
			}
		}()
	}
	wg.Wait()

	for i := 0; i < n*goroutines; i++ {
		if q.Dequeue() == nil {
			t.Errorf("Queue should have %d elements but got less", n*goroutines)
		}
	}

	if q.Dequeue() != nil {
		t.Errorf("Queue should be empty, but got more elements")
	}
}


func TestMultiGoroutineDequeue(t *testing.T) {
	q := NewLockFreeQueue()
	wg := &sync.WaitGroup{}
	n := 1000
	goroutines := 10
	var mu sync.Mutex
	seen := make(map[int]bool)

	for i := 0; i < n*goroutines; i++ {
		req := &Request{Id: i}
		q.Enqueue(req)
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				req := q.Dequeue()
				if req == nil {
					break
				}
				mu.Lock()
				if seen[req.Id] {
					t.Errorf("Dequeued duplicate request with ID: %d", req.Id)
				}
				seen[req.Id] = true
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(seen) != n*goroutines {
		t.Errorf("Should have dequeued %d unique items, but got %d", n*goroutines, len(seen))
	}
}

