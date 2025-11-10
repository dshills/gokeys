package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dshills/gokeys/input"
)

// Demo modes
type DemoMode int

const (
	ModeMenu DemoMode = iota
	ModeEventInspector
	ModeStateTracker
	ModeGameInput
	ModePerformance
	ModeUTF8Test
	ModeModifierTest
)

// Statistics for performance monitoring
type Stats struct {
	eventCount    int
	startTime     time.Time
	lastEventTime time.Time
	minLatency    time.Duration
	maxLatency    time.Duration
	totalLatency  time.Duration
}

func (s *Stats) update(event input.Event) {
	s.eventCount++
	s.lastEventTime = time.Now()

	if s.startTime.IsZero() {
		s.startTime = time.Now()
		s.minLatency = time.Hour // Start with large value
	}

	latency := time.Since(event.Timestamp)
	if latency < s.minLatency {
		s.minLatency = latency
	}
	if latency > s.maxLatency {
		s.maxLatency = latency
	}
	s.totalLatency += latency
}

func (s *Stats) avgLatency() time.Duration {
	if s.eventCount == 0 {
		return 0
	}
	return s.totalLatency / time.Duration(s.eventCount)
}

func (s *Stats) eventsPerSecond() float64 {
	elapsed := time.Since(s.startTime)
	if elapsed == 0 {
		return 0
	}
	return float64(s.eventCount) / elapsed.Seconds()
}

type AdvancedDemo struct {
	input        input.Input
	game         input.GameInput
	mode         DemoMode
	stats        Stats
	inputStarted bool
}

func NewAdvancedDemo() *AdvancedDemo {
	return &AdvancedDemo{
		input: input.New(),
		game:  input.NewGameInput(nil),
		mode:  ModeMenu,
	}
}

func (d *AdvancedDemo) ensureInputStopped() {
	if d.inputStarted {
		d.input.Stop()
		d.inputStarted = false
	}
}

func (d *AdvancedDemo) startInput() error {
	d.ensureInputStopped()
	if err := d.input.Start(); err != nil {
		return fmt.Errorf("failed to start input: %w", err)
	}
	d.inputStarted = true
	return nil
}

func (d *AdvancedDemo) stopInput() {
	d.ensureInputStopped()
}

func (d *AdvancedDemo) clearScreen() {
	fmt.Print("\033[2J\033[H") // ANSI clear screen and move cursor to top
}

func (d *AdvancedDemo) printHeader(title string) {
	d.clearScreen()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("  %s\n", title)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
}

func (d *AdvancedDemo) showMenu() {
	d.printHeader("GOKEYS ADVANCED DEMO - Main Menu")
	fmt.Println("Select a demo mode:")
	fmt.Println()
	fmt.Println("  [1] Event Inspector    - View detailed event information")
	fmt.Println("  [2] State Tracker       - Real-time key state monitoring")
	fmt.Println("  [3] Game Input Demo     - Action binding and game loop")
	fmt.Println("  [4] Performance Monitor - Latency and throughput stats")
	fmt.Println("  [5] UTF-8 Test          - Multi-byte character support")
	fmt.Println("  [6] Modifier Test       - Shift/Alt/Ctrl combinations")
	fmt.Println()
	fmt.Println("  [Q] Quit")
	fmt.Println()
	fmt.Println("Press a number key to select a mode, or Q to quit")
	fmt.Println(strings.Repeat("-", 80))
}

func (d *AdvancedDemo) runEventInspector() error {
	d.printHeader("EVENT INSPECTOR - Detailed Event Analysis")
	fmt.Println("This mode shows all event fields in detail")
	fmt.Println("Try different keys, combinations, and characters")
	fmt.Println("Press ESC to return to menu")
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))

	if err := d.startInput(); err != nil {
		return err
	}
	defer d.stopInput()

	for {
		event, ok := d.input.Poll()
		if !ok {
			return nil
		}

		if event.Key == input.KeyEscape {
			d.mode = ModeMenu
			return nil
		}

		d.stats.update(event)

		// Display event details
		fmt.Printf("\n[Event #%d]\n", d.stats.eventCount)
		fmt.Printf("  Key:       %v\n", event.Key)
		fmt.Printf("  Rune:      %q (U+%04X)\n", event.Rune, event.Rune)
		fmt.Printf("  Modifiers: %s\n", formatModifiers(event.Modifiers))
		fmt.Printf("  Pressed:   %v\n", event.Pressed)
		fmt.Printf("  Repeat:    %v\n", event.Repeat)
		fmt.Printf("  Timestamp: %v\n", event.Timestamp.Format("15:04:05.000"))
		fmt.Printf("  Latency:   %v\n", time.Since(event.Timestamp))
		fmt.Println(strings.Repeat("-", 80))
	}
}

func (d *AdvancedDemo) runStateTracker() error {
	d.printHeader("STATE TRACKER - Real-Time Key State Monitoring")
	fmt.Println("This mode tracks multiple key states simultaneously")
	fmt.Println("Try holding multiple keys and see their states")
	fmt.Println("Press ESC to return to menu")
	fmt.Println()

	if err := d.startInput(); err != nil {
		return err
	}
	defer d.stopInput()

	// Keys to track
	trackedKeys := []input.Key{
		input.KeyA, input.KeyS, input.KeyD, input.KeyW,
		input.KeyUp, input.KeyDown, input.KeyLeft, input.KeyRight,
		input.KeySpace, input.KeyEnter, input.KeyCtrlC,
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Display current state of all tracked keys
			d.clearScreen()
			d.printHeader("STATE TRACKER - Live Key States")

			fmt.Println("Currently Pressed Keys:")
			fmt.Println()

			anyPressed := false
			for _, key := range trackedKeys {
				if d.input.IsPressed(key) {
					fmt.Printf("  ✓ %v is PRESSED\n", key)
					anyPressed = true
				}
			}

			if !anyPressed {
				fmt.Println("  (no keys pressed)")
			}

			fmt.Println()
			fmt.Printf("Events processed: %d\n", d.stats.eventCount)
			fmt.Println()
			fmt.Println("Hold keys to see their state - ESC to return to menu")

		default:
			// Process events
			event := d.input.Next()
			if event == nil {
				time.Sleep(time.Millisecond)
				continue
			}

			d.stats.update(*event)

			if event.Key == input.KeyEscape {
				d.mode = ModeMenu
				return nil
			}
		}
	}
}

func (d *AdvancedDemo) runGameInput() error {
	d.printHeader("GAME INPUT - Action Mapping Demo")
	fmt.Println("This mode demonstrates game-style action binding")
	fmt.Println()

	if err := d.game.Start(); err != nil {
		return fmt.Errorf("failed to start game input: %w", err)
	}
	defer d.game.Stop()

	// P1: Basic bindings
	d.game.Bind("move-up", input.KeyW, input.KeyUp)
	d.game.Bind("move-down", input.KeyS, input.KeyDown)
	d.game.Bind("move-left", input.KeyA, input.KeyLeft)
	d.game.Bind("move-right", input.KeyD, input.KeyRight)

	// P2: Multiple keys per action
	d.game.Bind("jump", input.KeySpace, input.KeyJ)
	d.game.Bind("fire", input.KeyF, input.KeyEnter)
	d.game.Bind("special", input.KeyE, input.KeyZ)

	// P3: Menu actions
	d.game.Bind("pause", input.KeyP)
	d.game.Bind("rebind", input.KeyR)
	d.game.Bind("quit", input.KeyEscape, input.KeyQ)

	fmt.Println("CONTROLS:")
	fmt.Println("  Movement: WASD or Arrow Keys")
	fmt.Println("  Jump:     SPACE or J")
	fmt.Println("  Fire:     F or ENTER")
	fmt.Println("  Special:  E or Z")
	fmt.Println("  Pause:    P")
	fmt.Println("  Rebind:   R (rebinds jump to X)")
	fmt.Println("  Quit:     ESC or Q")
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))

	// Game state
	playerX, playerY := 40, 12
	rebound := false

	ticker := time.NewTicker(time.Second / 60) // 60 FPS
	defer ticker.Stop()

	lastRender := time.Now()

	for {
		<-ticker.C

		// Check actions
		if d.game.IsActionPressed("quit") {
			d.mode = ModeMenu
			return nil
		}

		if d.game.IsActionPressed("rebind") && !rebound {
			d.game.Bind("jump", input.KeyX) // Rebind to X only
			rebound = true
			fmt.Println("\n[REBIND] Jump action now bound to X only!")
		}

		if d.game.IsActionPressed("pause") {
			fmt.Println("\n[PAUSE] Game paused")
			time.Sleep(time.Second)
		}

		// Movement
		if d.game.IsActionPressed("move-left") {
			playerX--
			if playerX < 0 {
				playerX = 0
			}
		}
		if d.game.IsActionPressed("move-right") {
			playerX++
			if playerX > 79 {
				playerX = 79
			}
		}
		if d.game.IsActionPressed("move-up") {
			playerY--
			if playerY < 0 {
				playerY = 0
			}
		}
		if d.game.IsActionPressed("move-down") {
			playerY++
			if playerY > 20 {
				playerY = 20
			}
		}

		// Actions with visual feedback
		actionMsg := ""
		if d.game.IsActionPressed("jump") {
			actionMsg = "⬆ JUMP"
		} else if d.game.IsActionPressed("fire") {
			actionMsg = "💥 FIRE"
		} else if d.game.IsActionPressed("special") {
			actionMsg = "✨ SPECIAL"
		}

		// Render at 10 FPS to avoid flicker
		if time.Since(lastRender) > time.Second/10 {
			d.clearScreen()
			d.printHeader("GAME INPUT - Action Mapping Demo")

			// Draw game area
			for y := 0; y < 21; y++ {
				for x := 0; x < 80; x++ {
					if x == playerX && y == playerY {
						fmt.Print("@") // Player
					} else {
						fmt.Print(".")
					}
				}
				fmt.Println()
			}

			fmt.Println()
			fmt.Printf("Position: (%d, %d)  |  Action: %s\n", playerX, playerY, actionMsg)
			if rebound {
				fmt.Println("Status: Jump rebound to X key")
			}

			lastRender = time.Now()
		}
	}
}

func (d *AdvancedDemo) runPerformance() error {
	d.printHeader("PERFORMANCE MONITOR - Latency & Throughput Analysis")
	fmt.Println("This mode measures input system performance")
	fmt.Println("Type rapidly to see latency and throughput metrics")
	fmt.Println("Press ESC to return to menu")
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))

	if err := d.startInput(); err != nil {
		return err
	}
	defer d.stopInput()

	// Reset stats
	d.stats = Stats{}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update display
			fmt.Print("\033[8;0H") // Move cursor to line 8
			fmt.Println("PERFORMANCE METRICS:")
			fmt.Println()
			fmt.Printf("  Total Events:     %d\n", d.stats.eventCount)
			fmt.Printf("  Events/sec:       %.2f\n", d.stats.eventsPerSecond())
			fmt.Printf("  Min Latency:      %v\n", d.stats.minLatency)
			fmt.Printf("  Max Latency:      %v\n", d.stats.maxLatency)
			fmt.Printf("  Avg Latency:      %v\n", d.stats.avgLatency())
			fmt.Printf("  Uptime:           %v\n", time.Since(d.stats.startTime).Round(time.Second))
			fmt.Println()
			fmt.Println("Keep typing to update statistics...")

		default:
			event := d.input.Next()
			if event == nil {
				time.Sleep(time.Millisecond)
				continue
			}

			d.stats.update(*event)

			if event.Key == input.KeyEscape {
				d.mode = ModeMenu
				return nil
			}
		}
	}
}

func (d *AdvancedDemo) runUTF8Test() error {
	d.printHeader("UTF-8 TEST - Multi-byte Character Support")
	fmt.Println("This mode tests UTF-8 character input handling")
	fmt.Println("Try typing various characters:")
	fmt.Println("  - ASCII: a, b, c")
	fmt.Println("  - Extended ASCII: é, ñ, ü")
	fmt.Println("  - Emoji: 😀, 🎮, 🚀 (if your terminal supports it)")
	fmt.Println("  - CJK: 中, 日, 한")
	fmt.Println()
	fmt.Println("Press ESC to return to menu")
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))

	if err := d.startInput(); err != nil {
		return err
	}
	defer d.stopInput()

	buffer := ""

	for {
		event, ok := d.input.Poll()
		if !ok {
			return nil
		}

		if event.Key == input.KeyEscape {
			d.mode = ModeMenu
			return nil
		}

		if event.Key == input.KeyEnter {
			buffer += "\n"
		} else if event.Key == input.KeyBackspace && len(buffer) > 0 {
			// Remove last rune
			runes := []rune(buffer)
			buffer = string(runes[:len(runes)-1])
		} else if event.Rune != 0 {
			buffer += string(event.Rune)
		}

		d.stats.update(event)

		// Display
		fmt.Print("\033[11;0H") // Move cursor to line 11
		fmt.Println("INPUT BUFFER:")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println(buffer)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println()

		if event.Rune != 0 {
			fmt.Printf("Last character: %q (U+%04X) - %d bytes in UTF-8\n",
				event.Rune, event.Rune, len(string(event.Rune)))
		}

		fmt.Printf("Buffer length: %d characters, %d bytes\n",
			len([]rune(buffer)), len(buffer))
	}
}

func (d *AdvancedDemo) runModifierTest() error {
	d.printHeader("MODIFIER TEST - Shift/Alt/Ctrl Combinations")
	fmt.Println("This mode tests modifier key combinations")
	fmt.Println("Try pressing keys with modifiers:")
	fmt.Println("  - Shift + letter")
	fmt.Println("  - Ctrl + letter")
	fmt.Println("  - Alt + letter")
	fmt.Println("  - Combinations: Ctrl+Shift+A, etc.")
	fmt.Println()
	fmt.Println("Press ESC to return to menu")
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))

	if err := d.startInput(); err != nil {
		return err
	}
	defer d.stopInput()

	history := make([]string, 0, 10)

	for {
		event, ok := d.input.Poll()
		if !ok {
			return nil
		}

		if event.Key == input.KeyEscape {
			d.mode = ModeMenu
			return nil
		}

		d.stats.update(event)

		// Format event
		modStr := formatModifiers(event.Modifiers)
		keyStr := fmt.Sprintf("%v", event.Key)

		combo := modStr
		if combo != "" {
			combo += " + " + keyStr
		} else {
			combo = keyStr
		}

		if event.Rune != 0 {
			combo += fmt.Sprintf(" (%q)", event.Rune)
		}

		// Add to history
		history = append(history, combo)
		if len(history) > 10 {
			history = history[1:]
		}

		// Display
		fmt.Print("\033[11;0H") // Move cursor to line 11
		fmt.Println("RECENT COMBINATIONS:")
		fmt.Println(strings.Repeat("-", 80))
		for i := len(history) - 1; i >= 0; i-- {
			fmt.Printf("  %d. %s\n", len(history)-i, history[i])
		}
		fmt.Println(strings.Repeat("-", 80))
	}
}

func (d *AdvancedDemo) Run() error {
	for {
		switch d.mode {
		case ModeMenu:
			d.showMenu()
			if err := d.handleMenuInput(); err != nil {
				return err
			}

		case ModeEventInspector:
			if err := d.runEventInspector(); err != nil {
				return err
			}

		case ModeStateTracker:
			if err := d.runStateTracker(); err != nil {
				return err
			}

		case ModeGameInput:
			if err := d.runGameInput(); err != nil {
				return err
			}

		case ModePerformance:
			if err := d.runPerformance(); err != nil {
				return err
			}

		case ModeUTF8Test:
			if err := d.runUTF8Test(); err != nil {
				return err
			}

		case ModeModifierTest:
			if err := d.runModifierTest(); err != nil {
				return err
			}
		}
	}
}

func (d *AdvancedDemo) handleMenuInput() error {
	if err := d.startInput(); err != nil {
		return err
	}
	defer d.stopInput()

	for {
		event, ok := d.input.Poll()
		if !ok {
			return nil
		}

		switch event.Key {
		case input.Key1:
			d.mode = ModeEventInspector
			d.stats = Stats{} // Reset stats
			return nil
		case input.Key2:
			d.mode = ModeStateTracker
			d.stats = Stats{}
			return nil
		case input.Key3:
			d.mode = ModeGameInput
			d.stats = Stats{}
			return nil
		case input.Key4:
			d.mode = ModePerformance
			d.stats = Stats{}
			return nil
		case input.Key5:
			d.mode = ModeUTF8Test
			d.stats = Stats{}
			return nil
		case input.Key6:
			d.mode = ModeModifierTest
			d.stats = Stats{}
			return nil
		case input.KeyQ:
			return fmt.Errorf("quit")
		}
	}
}

func formatModifiers(mod input.Modifier) string {
	if mod == input.ModNone {
		return ""
	}

	parts := []string{}
	if mod&input.ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if mod&input.ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if mod&input.ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}

	return strings.Join(parts, "+")
}

func main() {
	demo := NewAdvancedDemo()

	if err := demo.Run(); err != nil {
		if err.Error() != "quit" {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("\nThank you for using the gokeys advanced demo!")
}
