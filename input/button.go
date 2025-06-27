package input

import (
	"fmt"
	"machine"
	"time"
	"esp32/hardware"
	"esp32/system"
	"esp32/audio"
	"esp32/display"
)

// Initialize button
func InitButton() {
	hardware.BootButton.Configure(machine.PinConfig{Mode: machine.PinInput})
}

// Check button press
func CheckButtonPress() {
	if !hardware.BootButton.Get() {
		time.Sleep(50 * time.Millisecond) // Debounce
		if !hardware.BootButton.Get() {
			HandleButtonPress()
			// Wait for button release
			for !hardware.BootButton.Get() {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

// Handle button press event
func HandleButtonPress() {
	fmt.Println("BOOT button press detected")

	// Play button confirmation sound
	audio.PlayBeep(1000, 100*time.Millisecond)

	switch system.CurrentState {
	case system.StateIdle:
		// Start recording
		audio.StartRecording()

	case system.StateRecording:
		// Stop recording
		audio.StopRecording()

	case system.StatePlaying:
		// Stop playing
		audio.StopPlaying()

	case system.StateProcessing:
		// Cancel processing
		system.CurrentState = system.StateIdle
		display.DisplayMessage("System", "Cancelled")
		hardware.StatusLED.Low()
	}
}