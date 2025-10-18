// Package main demonstrates basic usage of the gokeys input system.
// This example shows a simple event loop that captures keyboard input
// and displays key presses until the user presses Ctrl+C or 'q'.
package main

import (
	"fmt"
	"log"

	"github.com/dshills/gokeys/input"
)

func main() {
	// Create a new input system
	in := input.New()

	// Start the input system (enters raw mode)
	if err := in.Start(); err != nil {
		log.Fatalf("Failed to start input system: %v", err)
	}

	// Ensure terminal is restored on exit
	defer in.Stop()

	fmt.Println("gokeys basic example - Press keys to see events (Ctrl+C or 'q' to quit)")
	fmt.Println("---")

	// Main event loop using Poll (blocking)
	for {
		// Poll blocks until an event is available or system shuts down
		event, ok := in.Poll()
		if !ok {
			// System is shutting down
			break
		}

		// Display the event
		displayEvent(event)

		// Exit on Ctrl+C or 'q'
		if event.Key == input.KeyCtrlC || event.Rune == 'q' {
			fmt.Println("\nExiting...")
			break
		}
	}
}

// displayEvent prints information about a keyboard event
func displayEvent(e input.Event) {
	fmt.Printf("Key: %-15s", e.Key)

	if e.Rune != 0 {
		fmt.Printf(" Rune: %c", e.Rune)
	}

	if e.Modifiers != input.ModNone {
		fmt.Printf(" Mods: ")
		if e.Modifiers&input.ModShift != 0 {
			fmt.Print("Shift ")
		}
		if e.Modifiers&input.ModAlt != 0 {
			fmt.Print("Alt ")
		}
		if e.Modifiers&input.ModCtrl != 0 {
			fmt.Print("Ctrl ")
		}
	}

	if e.Repeat {
		fmt.Print(" [Repeat]")
	}

	fmt.Println()
}
