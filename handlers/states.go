package handlers

import (
	"esp32/audio"
	"esp32/display"
	"esp32/hardware"
	"esp32/system"
	"fmt"
	"time"
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

	// 读取真实的I2S音频数据
	buffer := make([]int16, 256)                       // 每次读取256个样本
	samplesRead, err := audio.ReadSamples(buffer, 100) // 100ms超时
	if err != nil {
		fmt.Printf("I2S read error: %v\n", err)
	} else if samplesRead > 0 {
		// 将读取的样本添加到音频缓冲区
		for i := 0; i < samplesRead && system.RecordedSamples < len(system.AudioBuffer); i++ {
			system.AudioBuffer[system.RecordedSamples] = buffer[i]
			system.RecordedSamples++
		}
	}

	// Update display
	display.DisplayRecordingStatus()

	// Check if maximum recording time is reached
	if system.RecordingTime >= hardware.MAX_RECORD_TIME {
		audio.StopRecording()
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
		audio.StopPlaying()
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
