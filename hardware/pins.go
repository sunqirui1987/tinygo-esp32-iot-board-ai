package hardware

import "machine"

// Pin definitions
var (
	// LED indicator
	StatusLED = machine.GPIO2

	// Boot button (BOOT BUTTON)
	BootButton = machine.GPIO34 // A6 pin

	// OLED display I2C pins
	SclPin = machine.GPIO22 // OLED SCL pin
	SdaPin = machine.GPIO21 // OLED SDA pin

	// Microphone pins (INMP441)
	MicDINPin = machine.GPIO27 // Microphone DIN pin
	MicWSPin  = machine.GPIO26 // Microphone WS pin
	MicSCKPin = machine.GPIO25 // Microphone SCK pin

	// Speaker pins
	SpeakerDOUTPin = machine.GPIO23 // Speaker DOUT pin
	SpeakerBCLKPin = machine.GPIO33 // Speaker BCLK pin
	SpeakerWSPin   = machine.GPIO32 // Speaker WS pin
)

// Audio configuration
const (
	SAMPLE_RATE     = 16000 // 16kHz sample rate
	BITS_PER_SAMPLE = 16    // 16-bit sampling
	MAX_RECORD_TIME = 10    // Maximum recording time (seconds)
	BUFFER_SIZE     = 1024  // Audio buffer size
)