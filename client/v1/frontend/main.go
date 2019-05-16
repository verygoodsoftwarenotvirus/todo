// +build wasm

package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("getting body")

	// suspend loop
	for {
		select {
		case <-time.NewTicker((1<<31 - 1) * time.Second).C:
			//
		}
	}
}
