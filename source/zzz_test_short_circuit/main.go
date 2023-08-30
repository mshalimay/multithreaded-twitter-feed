package main

import (
	"fmt"
	"math"
)


func main() {
	x :=  1 << 31 - 1;
	fmt.Println(x)
	fmt.Println(math.MaxInt32)
    for i := 0; i < 10; i++ {
        if !testFunc(1) && !testFunc(2) {
            // do nothing
        }
    }
}

func testFunc(i int) bool {
    fmt.Printf("function %d called\n", i)
    return true
}