package main

import (
	"fmt"
	"proj2/lock"
	"sync"
	"time"
	"math/rand"
)

func randomT(minTime int, maxTime int) time.Duration {
	seconds := minTime + rand.Intn(maxTime-1) 
	return time.Duration(seconds) * time.Second
}

func readFunc(rw lock.RWLock, wg *sync.WaitGroup, minTime int, maxTime int) {
	defer wg.Done()
	rw.RLock()
	fmt.Printf("Reader %d is reading\n", lock.GetGID())
	time.Sleep(randomT(minTime, maxTime)) // simulate read
	fmt.Printf("Reader %d is done reading\n", lock.GetGID())
	rw.RUnlock()
}

func writeFunc(rw lock.RWLock, wg *sync.WaitGroup, minTime int, maxTime int) {
	defer wg.Done()
	rw.Lock()
	fmt.Printf("Writer %d is writing\n", lock.GetGID())
	time.Sleep(randomT(minTime, maxTime)) // simulate write
	fmt.Printf("Writer %d is done writing\n", lock.GetGID())
	rw.Unlock()
}



func main() {
	rand.Seed(time.Now().UnixNano())
	// rand.Seed(0)
	
	start := time.Now()

	var readWg, writeWg sync.WaitGroup
	rw := lock.NewRWLock()
	// rw := lock.NewRMWAtomicLock()
	
	writerSpawned := 0
	writerMax := 2
	for i := 0; i < 200; i++ {
		spawnWriter := rand.Intn(5) == 0
		if spawnWriter && writerSpawned < writerMax {
			writeWg.Add(1)
			writerSpawned++
			go writeFunc(rw, &writeWg, 1, 3)
		
		} else{
			readWg.Add(1)
			go readFunc(rw, &readWg, 1, 3)
		}
	}

	readWg.Wait()
	writeWg.Wait()

	fmt.Printf("Time elapsed: %v\n", time.Since(start))
}
