// Package main demonstrates the Poll() API for blocking event retrieval.
// Poll() is ideal for applications that want to wait for user input
// without consuming CPU in a busy loop.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dshills/gokeys/input"
)

func main() {
	in := input.New()

	if err := in.Start(); err != nil {
		log.Fatalf("Failed to start input system: %v", err)
	}
	defer in.Stop()

	fmt.Println("gokeys Poll() example - Demonstrates blocking event retrieval")
	fmt.Println("Press any key (Escape to quit)")
	fmt.Println("---")

	eventCount := 0
	startTime := time.Now()

	// Poll() blocks until an event is available
	// This is efficient - no CPU wasted in busy loops
	for {
		event, ok := in.Poll()
		if !ok {
			// System is shutting down
			break
		}

		eventCount++

		// Display event with timing information
		elapsed := time.Since(startTime)
		fmt.Printf("[%8s] Event #%d: Key=%v", elapsed.Round(time.Millisecond), eventCount, event.Key)

		if event.Rune != 0 {
			fmt.Printf(" ('%c')", event.Rune)
		}

		// Show that IsPressed() reflects current state
		if in.IsPressed(event.Key) {
			fmt.Print(" [Currently Pressed]")
		}

		fmt.Println()

		// Exit on Escape key
		if event.Key == input.KeyEscape {
			fmt.Println("\nEscape pressed - exiting...")
			break
		}
	}

	// Show statistics
	elapsed := time.Since(startTime)
	fmt.Printf("\nTotal events: %d in %v\n", eventCount, elapsed.Round(time.Millisecond))
}
