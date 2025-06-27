package audio

import (
	"fmt"
	"machine"
	"time"
	"esp32/hardware"
	"esp32/system"
	"esp32/display"
)

// Initialize I2S audio interface
func InitI2S() {
	fmt.Println("Initializing I2S audio interface...")

	// Configure microphone pins
	hardware.MicDINPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	hardware.MicWSPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hardware.MicSCKPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Configure speaker pins
	hardware.SpeakerDOUTPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hardware.SpeakerBCLKPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hardware.SpeakerWSPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Initialize I2S configuration
	system.I2sInitialized = true
	fmt.Println("I2S audio interface initialization completed")
}

// Start recording
func StartRecording() {
	if !system.I2sInitialized {
		return
	}

	system.CurrentState = system.StateRecording
	system.RecordingTime = 0.0
	system.RecordedSamples = 0
	hardware.StatusLED.High()

	display.DisplayRecordingStatus()
	fmt.Println("Recording started...")
}

// Stop recording
func StopRecording() {
	if system.CurrentState != system.StateRecording {
		return
	}

	system.CurrentState = system.StateProcessing
	hardware.StatusLED.Low()

	fmt.Printf("Recording completed, duration: %.1f seconds\n", system.RecordingTime)
	display.DisplayMessage("Recording Done", fmt.Sprintf("%.1fs Press to play", system.RecordingTime))

	// Brief processing then enter idle state
	time.Sleep(1 * time.Second)
	system.CurrentState = system.StateIdle
}

// Start playing
func StartPlaying() {
	if system.RecordedSamples == 0 {
		display.DisplayMessage("Error", "No recording")
		return
	}

	system.CurrentState = system.StatePlaying
	system.PlayingTime = 0.0
	system.PlaybackPosition = 0
	hardware.StatusLED.High()

	display.DisplayPlayingStatus()
	fmt.Println("Playing recording...")
}

// Stop playing
func StopPlaying() {
	if system.CurrentState != system.StatePlaying {
		return
	}

	system.CurrentState = system.StateIdle
	hardware.StatusLED.Low()

	fmt.Println("Playback stopped")
	display.DisplayMessage("Playback Stop", "Press to replay")
}

// Play beep sound
func PlayBeep(frequency int, duration time.Duration) {
	period := time.Duration(1000000/frequency) * time.Microsecond
	half := period / 2

	start := time.Now()
	for time.Since(start) < duration {
		hardware.SpeakerDOUTPin.High()
		time.Sleep(half)
		hardware.SpeakerDOUTPin.Low()
		time.Sleep(half)
	}
}