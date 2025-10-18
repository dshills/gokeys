// Package main demonstrates the Next() API for non-blocking event retrieval.
// Next() is ideal for game loops and applications that need to do work
// between checking for input events.
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

	fmt.Println("gokeys Next() example - Demonstrates non-blocking event retrieval")
	fmt.Println("This simulates a game loop that runs at 60 FPS")
	fmt.Println("Press keys to see events (Escape to quit)")
	fmt.Println("---")

	const targetFPS = 60
	const frameDuration = time.Second / targetFPS

	frameCount := 0
	running := true
	startTime := time.Now()

	// Game loop
	for running {
		frameStart := time.Now()

		// Process all available events (non-blocking)
		eventsThisFrame := 0
		for {
			event := in.Next()
			if event == nil {
				// No more events available
				break
			}

			eventsThisFrame++

			// Handle the event
			handleEvent(*event, &running)
		}

		// Do game update work (simulated)
		frameCount++

		// Display frame info every second
		if frameCount%targetFPS == 0 {
			elapsed := time.Since(startTime)
			fps := float64(frameCount) / elapsed.Seconds()
			fmt.Printf("Frame %d: %.1f FPS, %d events this frame\n", frameCount, fps, eventsThisFrame)
		}

		// Sleep to maintain target FPS
		frameDuration := time.Since(frameStart)
		if frameDuration < time.Second/targetFPS {
			time.Sleep(time.Second/targetFPS - frameDuration)
		}
	}

	elapsed := time.Since(startTime)
	actualFPS := float64(frameCount) / elapsed.Seconds()
	fmt.Printf("\nRan %d frames in %v (%.1f FPS)\n", frameCount, elapsed.Round(time.Millisecond), actualFPS)
}

func handleEvent(event input.Event, running *bool) {
	fmt.Printf("  Event: Key=%v", event.Key)

	if event.Rune != 0 {
		fmt.Printf(" ('%c')", event.Rune)
	}

	if event.Modifiers != input.ModNone {
		fmt.Print(" [")
		if event.Modifiers&input.ModShift != 0 {
			fmt.Print("Shift ")
		}
		if event.Modifiers&input.ModAlt != 0 {
			fmt.Print("Alt ")
		}
		if event.Modifiers&input.ModCtrl != 0 {
			fmt.Print("Ctrl")
		}
		fmt.Print("]")
	}

	if event.Repeat {
		fmt.Print(" [Repeat]")
	}

	fmt.Println()

	// Exit on Escape
	if event.Key == input.KeyEscape {
		fmt.Println("  Escape pressed - exiting...")
		*running = false
	}
}
