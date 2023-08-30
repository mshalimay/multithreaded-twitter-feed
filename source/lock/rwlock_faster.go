// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.
package lock

import (
	"sync"
)


// rwlock is the internal representation of a r/w lock
type rwLockFaster struct {
	mutex 	 		*sync.Mutex
	rCond        	*sync.Cond		// condition variable for waiting readers
	wCond        	*sync.Cond		// condition variable for waiting writers
	w2Cond        	*sync.Cond		// condition variable for writer waiting readers to finish
	
	readingReaders 	int				// # readers reading
	pendingReaders 	int				// # readers that were reading or waiting when a writer arrived
	
	writerWaiting 	bool			// signals if a writer is waiting
	writerWriting 	bool			// signals if a writer is writing
	waitForReaders 	int				// # readers that a writer is waiting for
	
}

// NewRWLock creates and returns a new r/w lock
func NewRWLockFaster() RWLock {
	var mutex sync.Mutex
	// dMutex := dummyLocker{}
	rCondVar := sync.NewCond(&mutex)
	wCondVar := sync.NewCond(&mutex)
	w2CondVar := sync.NewCond(&mutex)

	return &rwLockFaster{mutex: &mutex, rCond: rCondVar, wCond: wCondVar, w2Cond: w2CondVar,
		readingReaders: 0, pendingReaders: 0, writerWaiting: false, writerWriting: false, waitForReaders: 0}
}

// RLock acuires a reader if there is no writer using the lock and if there are less than `maxReaders` readers
func (rw *rwLockFaster) RLock() {
	rw.mutex.Lock()

	// if writer is writing, put itself in the queue
	for rw.writerWriting || rw.writerWaiting {
		rw.rCond.Wait()		
	}

	// if more than `maxReaders` pending readers, update pending readers and wait
	for rw.readingReaders > maxReaders {
		rw.pendingReaders++
		rw.rCond.Wait()
		rw.pendingReaders--
	}

	// update # active readers when unblocked
	rw.readingReaders++
	rw.mutex.Unlock()
}

// RUnlock reader unlock
func (rw *rwLockFaster) RUnlock() {
	rw.mutex.Lock()

	// finished reading -> update # active readers
	rw.readingReaders--
	
	// if no writers waiting and there are pending readers, wake up a reader
	if !rw.writerWaiting && rw.pendingReaders > 0{
		rw.rCond.Signal()
	
	// else, try to wake a writer
	} else {		
		rw.RUnlockSlow()
	}
	rw.mutex.Unlock()
}
	
func (rw *rwLockFaster) RUnlockSlow() {

	// if the waiting writer is waiting for readers, wake up a reader
	
	if rw.waitForReaders > 1 {
		rw.waitForReaders--
		rw.rCond.Signal()
	// the last reader that finishes reading wake up the writer
	} else {
		rw.w2Cond.Signal()
	}
}

// Lock writer lock: allows only one writer at a time to enter in critical section
func (rw *rwLockFaster) Lock() {
	rw.mutex.Lock()

	// if other writers writing, wait
	for rw.writerWaiting || rw.writerWriting {
		rw.wCond.Wait()
	}
	
	// if there are readers reading, wait
	if rw.readingReaders > 0 || rw.pendingReaders > 0 {
		// signalize that a writer is waiting
		rw.writerWaiting = true
		// take note of how many readers to wait for at this point in time
		rw.waitForReaders = rw.readingReaders + rw.pendingReaders
		// wait for readers to finish
		rw.w2Cond.Wait()
	}

	// update signals when unblocked
	rw.writerWriting = true
	rw.writerWaiting = false
	rw.mutex.Unlock()
}

// Unlock -> writer unlock the r/w lock
func (rw *rwLockFaster) Unlock() {
	rw.mutex.Lock()

	// signalize that writer is not writing anymore
	rw.writerWriting = false

	// wake up readers that came in while the writer was writing
	// obs: readers that came after writers waiting will do work before them.
	// This is the same behavior as the `Go's` RWLock implementation
	rw.rCond.Broadcast()

	// allow next writers to proceed, if any
	rw.wCond.Signal()
	rw.mutex.Unlock()
}

// DummyLocker implements the Locker interface but does nothing. It is useful when the lock is not needed,
// but we want to use condition variables for signaling. It is used in the consumer-producer in `server.go`.
type DummyLocker struct{
}

func (d *DummyLocker) Lock() {
}

func (d *DummyLocker) Unlock() {
}
