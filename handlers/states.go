package handlers

import (
	"time"
	"esp32/hardware"
	"esp32/system"
	"esp32/display"
)

// Handle idle state
func HandleIdleState() {
	// LED slow blink indicates idle
	hardware.StatusLED.High()
	time.Sleep(100 * time.Millisecond)
	hardware.StatusLED.Low()
	time.Sleep(1900 * time.Millisecond)

	// Display idle information
	if system.RecordedSamples > 0 {
		display.DisplayMessage("Ready", "Press to play")
	} else {
		display.DisplayMessage("Ready", "Press to record")
	}
}

// Handle recording state
func HandleRecordingState() {
	// LED fast blink indicates recording
	hardware.StatusLED.High()
	time.Sleep(200 * time.Millisecond)
	hardware.StatusLED.Low()
	time.Sleep(200 * time.Millisecond)

	// Update recording time
	system.RecordingTime += 0.4 // Update every 400ms

	// Simulate recording data collection
	if system.I2sInitialized {
		samplesThisCycle := int(hardware.SAMPLE_RATE * 0.4) // 400ms worth of samples
		for i := 0; i < samplesThisCycle && system.RecordedSamples < len(system.AudioBuffer); i++ {
			system.AudioBuffer[system.RecordedSamples] = int16((system.RecordedSamples % 1000) - 500)
			system.RecordedSamples++
		}
	}

	// Update display
	display.DisplayRecordingStatus()

	// Check if maximum recording time is reached
	if system.RecordingTime >= hardware.MAX_RECORD_TIME {
		// Stop recording (需要导入audio包)
	}
}

// Handle playing state
func HandlePlayingState() {
	// LED solid indicates playing
	hardware.StatusLED.High()

	// Update playing time
	system.PlayingTime += 0.1 // Update every 100ms

	// Simulate playback data output
	if system.I2sInitialized {
		samplesThisCycle := int(hardware.SAMPLE_RATE * 0.1) // 100ms worth of samples
		system.PlaybackPosition += samplesThisCycle
	}

	// Update display
	display.DisplayPlayingStatus()

	// Check if playback is complete
	totalPlayTime := float64(system.RecordedSamples) / hardware.SAMPLE_RATE
	if system.PlayingTime >= totalPlayTime {
		// Stop playing (需要导入audio包)
	}
}

// Handle processing state
func HandleProcessingState() {
	// LED blink indicates processing
	for i := 0; i < 3; i++ {
		hardware.StatusLED.High()
		time.Sleep(100 * time.Millisecond)
		hardware.StatusLED.Low()
		time.Sleep(100 * time.Millisecond)
	}
}