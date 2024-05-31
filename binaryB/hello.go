package main

import (
	"fmt"
	"time"
)

func main() {
	counter := 0
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		counter++
		fmt.Printf("Hello, World2! Counter: %d\n", counter)
	}
}
