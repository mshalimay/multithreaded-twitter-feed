// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.
package lock

import (
	"sync"
	"runtime"
	"bytes"
	"strconv"
)

// maxReaders is the maximum number of readers allowed to enter the critical section
const maxReaders = 32;

// RWLock represents a reader/writer lock
type RWLock interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// rwlock is the internal representation of a r/w lock
type rwLock struct {
	mutex       *sync.Mutex
	cond        *sync.Cond
	readerCount int
	writerCount int
	writerWait  bool
}

// NewRWLock creates and returns a new r/w lock
func NewRWLock() RWLock {
	var mutex sync.Mutex
	condVar := sync.NewCond(&mutex)
	return &rwLock{mutex: &mutex, cond: condVar}
}

// RLock acuires a reader if there is no writer using the lock and if there are less than `maxReaders` readers
func (rw *rwLock) RLock() {
	rw.mutex.Lock()

	// if writer is waiting, wait
	for rw.writerCount > 0 {
		rw.cond.Wait()
	}

	// if more than `maxReaders`, wait
	for rw.readerCount > maxReaders {
		rw.cond.Wait()
	}

	rw.readerCount++
	rw.mutex.Unlock()
}

// RUnlock unlocks the reader lock
func (rw *rwLock) RUnlock() {
	rw.mutex.Lock()

	rw.readerCount--

	// when no readers, wake up all sleeping threads
	if rw.readerCount == 0 {
		rw.cond.Broadcast()
	}
	rw.mutex.Unlock()
}


// Lock writer lock: allows only one writer at a time to enter in critical section
func (rw *rwLock) Lock() {
	rw.mutex.Lock()
	rw.writerCount++

	// if other readers reading, wait
	for rw.readerCount > 0 {
		rw.cond.Wait()
	}
	// if there is a writer using, wait
	for rw.writerWait {
		rw.cond.Wait()
	}

	// obtain the lock
	rw.writerWait = true
	rw.mutex.Unlock()
}

// Unlock -> writer unlock the r/w lock
func (rw *rwLock) Unlock() {
	rw.mutex.Lock()

	// release the lock
	rw.writerCount--
	rw.writerWait = false

	// wake up sleeping threads
	rw.cond.Broadcast()

	rw.mutex.Unlock()
}


// GetGID returns the goroutine id of the caller
func GetGID() uint64 {
    b := make([]byte, 64)
    b = b[:runtime.Stack(b, false)]
    b = bytes.TrimPrefix(b, []byte("goroutine "))
    b = b[:bytes.IndexByte(b, ' ')]
    n, _ := strconv.ParseUint(string(b), 10, 64)
    return n
}


