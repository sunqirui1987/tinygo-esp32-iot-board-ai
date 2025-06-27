package system

import "tinygo.org/x/drivers/ssd1306"

// System state
type SystemState int

const (
	StateIdle SystemState = iota
	StateRecording
	StatePlaying
	StateProcessing
)

// Global system state
var (
	CurrentState     = StateIdle
	RecordingTime    = 0.0 // Recording duration (seconds)
	PlayingTime      = 0.0 // Playing duration (seconds)
	Display          ssd1306.Device
	OledInitialized  = false
	AudioBuffer      = make([]int16, 16000*10) // Audio buffer
	RecordedSamples  = 0                       // Recorded samples count
	PlaybackPosition = 0                       // Playback position
	I2sInitialized   = false
)