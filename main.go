package main

import (
	"fmt"
	"machine"
	"time"

	"esp32/hardware"
	"esp32/system"
	"esp32/display"
	"esp32/audio"
	"esp32/input"
	"esp32/handlers"
)

func main() {
	fmt.Println("ESP32 Recording Playback System Starting...")

	// Initialize hardware
	initHardware()

	// Display startup information
	display.DisplayMessage("ESP32", "Audio System Ready")
	time.Sleep(2 * time.Second)

	// Main loop
	for {
		// Check button status
		input.CheckButtonPress()

		// Execute corresponding operations based on current state
		switch system.CurrentState {
		case system.StateIdle:
			handlers.HandleIdleState()
		case system.StateRecording:
			handlers.HandleRecordingState()
		case system.StatePlaying:
			handlers.HandlePlayingState()
		case system.StateProcessing:
			handlers.HandleProcessingState()
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// Initialize hardware
func initHardware() {
	fmt.Println("Initializing hardware...")

	// Configure LED
	hardware.StatusLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hardware.StatusLED.Low()

	// Initialize components
	input.InitButton()
	display.InitOLED()
	audio.InitI2S()

	fmt.Println("Hardware initialization completed")
}
