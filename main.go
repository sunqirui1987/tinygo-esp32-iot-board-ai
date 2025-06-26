package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
)

// Pin definitions
var (
	// LED indicator
	statusLED = machine.GPIO2

	// Boot button (BOOT BUTTON)
	bootButton = machine.GPIO34 // A6 pin

	// OLED display I2C pins
	sclPin = machine.GPIO22 // OLED SCL pin
	sdaPin = machine.GPIO21 // OLED SDA pin

	// Microphone pins (INMP441)
	micDINPin = machine.GPIO27 // Microphone DIN pin
	micWSPin  = machine.GPIO26 // Microphone WS pin
	micSCKPin = machine.GPIO25 // Microphone SCK pin

	// Speaker pins
	speakerDOUTPin = machine.GPIO23 // Speaker DOUT pin
	speakerBCLKPin = machine.GPIO33 // Speaker BCLK pin
	speakerWSPin   = machine.GPIO32 // Speaker WS pin
)

// System state
type SystemState int

const (
	StateIdle SystemState = iota
	StateRecording
	StatePlaying
	StateProcessing
)

// Audio configuration
const (
	SAMPLE_RATE     = 16000 // 16kHz sample rate
	BITS_PER_SAMPLE = 16    // 16-bit sampling
	MAX_RECORD_TIME = 10    // Maximum recording time (seconds)
	BUFFER_SIZE     = 1024  // Audio buffer size
)

// Global variables
var (
	currentState     = StateIdle
	recordingTime    = 0.0 // Recording duration (seconds)
	playingTime      = 0.0 // Playing duration (seconds)
	display          ssd1306.Device
	oledInitialized  = false
	audioBuffer      = make([]int16, SAMPLE_RATE*MAX_RECORD_TIME) // Audio buffer
	recordedSamples  = 0                                          // Recorded samples count
	playbackPosition = 0                                          // Playback position
	i2sInitialized   = false
)

func main() {
	fmt.Println("ESP32 Recording Playback System Starting...")

	// Initialize hardware
	initHardware()

	// Display startup information
	displayMessage("ESP32", "Audio System Ready")
	time.Sleep(2 * time.Second)

	// Main loop
	for {
		// Check button status
		checkButtonPress()

		// Execute corresponding operations based on current state
		switch currentState {
		case StateIdle:
			handleIdleState()
		case StateRecording:
			handleRecordingState()
		case StatePlaying:
			handlePlayingState()
		case StateProcessing:
			handleProcessingState()
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// Initialize hardware
func initHardware() {
	fmt.Println("Initializing hardware...")

	// Configure LED
	statusLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	statusLED.Low()

	// Configure button
	bootButton.Configure(machine.PinConfig{Mode: machine.PinInput})

	// Initialize OLED display
	initOLED()

	// Initialize I2S audio interface
	initI2S()

	fmt.Println("Hardware initialization completed")
}

// Initialize OLED display
func initOLED() {
	fmt.Println("Initializing OLED display...")

	// Configure I2C
	machine.I2C0.Configure(machine.I2CConfig{
		SCL:       sclPin,
		SDA:       sdaPin,
		Frequency: 400000, // 400kHz
	})

	// Create SSD1306 device instance
	display = ssd1306.NewI2C(machine.I2C0)

	// Configure display
	display.Configure(ssd1306.Config{
		Address: ssd1306.Address_128_32, // 0x3C
		Width:   128,
		Height:  64,
	})

	// Clear display
	display.ClearDisplay()
	oledInitialized = true
	fmt.Println("OLED display initialization completed")
}

// Initialize I2S audio interface
func initI2S() {
	fmt.Println("Initializing I2S audio interface...")

	// Configure microphone pins
	micDINPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	micWSPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	micSCKPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Configure speaker pins
	speakerDOUTPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	speakerBCLKPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	speakerWSPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Initialize I2S configuration (simplified version, actual implementation needs specific hardware configuration)
	i2sInitialized = true
	fmt.Println("I2S audio interface initialization completed")
}

// Check button press
func checkButtonPress() {
	if !bootButton.Get() {
		time.Sleep(50 * time.Millisecond) // Debounce
		if !bootButton.Get() {
			handleButtonPress()
			// Wait for button release
			for !bootButton.Get() {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

// Handle button press event
func handleButtonPress() {
	fmt.Println("BOOT button press detected")

	// Play button confirmation sound
	playBeep(1000, 100*time.Millisecond)

	switch currentState {
	case StateIdle:
		// Start recording
		startRecording()

	case StateRecording:
		// Stop recording
		stopRecording()

	case StatePlaying:
		// Stop playing
		stopPlaying()

	case StateProcessing:
		// Cancel processing
		currentState = StateIdle
		displayMessage("System", "Cancelled")
		statusLED.Low()
	}
}

// Start recording
func startRecording() {
	if !i2sInitialized {
		return
	}

	currentState = StateRecording
	recordingTime = 0.0
	recordedSamples = 0
	statusLED.High()

	displayRecordingStatus()
	fmt.Println("Recording started...")
}

// Stop recording
func stopRecording() {
	if currentState != StateRecording {
		return
	}

	currentState = StateProcessing
	statusLED.Low()

	fmt.Printf("Recording completed, duration: %.1f seconds\n", recordingTime)
	displayMessage("Recording Done", fmt.Sprintf("%.1fs Press to play", recordingTime))

	// Brief processing then enter idle state
	time.Sleep(1 * time.Second)
	currentState = StateIdle
}

// Start playing
func startPlaying() {
	if recordedSamples == 0 {
		displayMessage("Error", "No recording")
		return
	}

	currentState = StatePlaying
	playingTime = 0.0
	playbackPosition = 0
	statusLED.High()

	displayPlayingStatus()
	fmt.Println("Playing recording...")
}

// Stop playing
func stopPlaying() {
	if currentState != StatePlaying {
		return
	}

	currentState = StateIdle
	statusLED.Low()

	fmt.Println("Playback stopped")
	displayMessage("Playback Stop", "Press to replay")
}

// Handle idle state
func handleIdleState() {
	// LED slow blink indicates idle
	statusLED.High()
	time.Sleep(100 * time.Millisecond)
	statusLED.Low()
	time.Sleep(1900 * time.Millisecond)

	// Display idle information
	if recordedSamples > 0 {
		displayMessage("Ready", "Press to play")
	} else {
		displayMessage("Ready", "Press to record")
	}
}

// Handle recording state
func handleRecordingState() {
	// LED fast blink indicates recording
	statusLED.High()
	time.Sleep(200 * time.Millisecond)
	statusLED.Low()
	time.Sleep(200 * time.Millisecond)

	// Update recording time
	recordingTime += 0.4 // Update every 400ms

	// Simulate recording data collection
	if i2sInitialized {
		// Should read actual audio data from I2S interface here
		// Now using simulated data instead
		samplesThisCycle := int(SAMPLE_RATE * 0.4) // 400ms worth of samples
		for i := 0; i < samplesThisCycle && recordedSamples < len(audioBuffer); i++ {
			// Simulated audio data (should read from microphone in reality)
			audioBuffer[recordedSamples] = int16((recordedSamples % 1000) - 500)
			recordedSamples++
		}
	}

	// Update display
	displayRecordingStatus()

	// Check if maximum recording time is reached
	if recordingTime >= MAX_RECORD_TIME {
		stopRecording()
	}
}

// Handle playing state
func handlePlayingState() {
	// LED solid indicates playing
	statusLED.High()

	// Update playing time
	playingTime += 0.1 // Update every 100ms

	// Simulate playback data output
	if i2sInitialized {
		// Should output audio data to I2S interface here
		// Now using simulated playback instead
		samplesThisCycle := int(SAMPLE_RATE * 0.1) // 100ms worth of samples
		playbackPosition += samplesThisCycle
	}

	// Update display
	displayPlayingStatus()

	// Check if playback is complete
	totalPlayTime := float64(recordedSamples) / SAMPLE_RATE
	if playingTime >= totalPlayTime {
		stopPlaying()
	}
}

// Handle processing state
func handleProcessingState() {
	// LED blink indicates processing
	for i := 0; i < 3; i++ {
		statusLED.High()
		time.Sleep(100 * time.Millisecond)
		statusLED.Low()
		time.Sleep(100 * time.Millisecond)
	}
}

// Display recording status
func displayRecordingStatus() {
	if !oledInitialized {
		return
	}

	display.ClearBuffer()
	drawSimpleText("Recording...", 0, 0)
	drawSimpleText(fmt.Sprintf("Time: %.1fs", recordingTime), 0, 20)
	drawSimpleText(fmt.Sprintf("Max: %ds", MAX_RECORD_TIME), 0, 40)
	display.Display()
}

// Display playing status
func displayPlayingStatus() {
	if !oledInitialized {
		return
	}

	totalTime := float64(recordedSamples) / SAMPLE_RATE
	display.ClearBuffer()
	drawSimpleText("> PLAY", 0, 0)
	drawSimpleText(fmt.Sprintf("%.1f/%.1fs", playingTime, totalTime), 0, 20)

	// Display progress bar
	progress := int(playingTime / totalTime * 120)
	if progress > 120 {
		progress = 120
	}
	for i := 0; i < progress; i++ {
		display.SetPixel(int16(i+4), 45, color.RGBA{255, 255, 255, 255})
	}

	display.Display()
}

// Display message to screen
func displayMessage(title, message string) {
	if !oledInitialized {
		return
	}

	display.ClearBuffer()
	drawSimpleText(title, 0, 0)
	drawSimpleText(message, 0, 20)
	display.Display()

	fmt.Printf("[Display] %s: %s\n", title, message)
}

// Draw simple text (using pixel matrix)
func drawSimpleText(text string, x, y int16) {
	charWidth := int16(6) // Including spacing
	currentX := x

	for _, char := range text {
		if currentX > 122 { // Prevent exceeding screen boundaries
			break
		}
		drawChar5x7(char, currentX, y)
		currentX += charWidth
	}
}

// Draw 5x7 pixel character - optimized complete character set
func drawChar5x7(char rune, x, y int16) {
	var pattern []uint8

	switch char {
	// Uppercase letters
	case 'A':
		pattern = []uint8{0x7C, 0x12, 0x11, 0x12, 0x7C}
	case 'B':
		pattern = []uint8{0x7F, 0x49, 0x49, 0x49, 0x36}
	case 'C':
		pattern = []uint8{0x3E, 0x41, 0x41, 0x41, 0x22}
	case 'D':
		pattern = []uint8{0x7F, 0x41, 0x41, 0x22, 0x1C}
	case 'E':
		pattern = []uint8{0x7F, 0x49, 0x49, 0x49, 0x41}
	case 'F':
		pattern = []uint8{0x7F, 0x09, 0x09, 0x09, 0x01}
	case 'G':
		pattern = []uint8{0x3E, 0x41, 0x49, 0x49, 0x7A}
	case 'H':
		pattern = []uint8{0x7F, 0x08, 0x08, 0x08, 0x7F}
	case 'I':
		pattern = []uint8{0x00, 0x41, 0x7F, 0x41, 0x00}
	case 'J':
		pattern = []uint8{0x20, 0x40, 0x41, 0x3F, 0x01}
	case 'K':
		pattern = []uint8{0x7F, 0x08, 0x14, 0x22, 0x41}
	case 'L':
		pattern = []uint8{0x7F, 0x40, 0x40, 0x40, 0x40}
	case 'M':
		pattern = []uint8{0x7F, 0x02, 0x0C, 0x02, 0x7F}
	case 'N':
		pattern = []uint8{0x7F, 0x04, 0x08, 0x10, 0x7F}
	case 'O':
		pattern = []uint8{0x3E, 0x41, 0x41, 0x41, 0x3E}
	case 'P':
		pattern = []uint8{0x7F, 0x09, 0x09, 0x09, 0x06}
	case 'Q':
		pattern = []uint8{0x3E, 0x41, 0x51, 0x21, 0x5E}
	case 'R':
		pattern = []uint8{0x7F, 0x09, 0x19, 0x29, 0x46}
	case 'S':
		pattern = []uint8{0x46, 0x49, 0x49, 0x49, 0x31}
	case 'T':
		pattern = []uint8{0x01, 0x01, 0x7F, 0x01, 0x01}
	case 'U':
		pattern = []uint8{0x3F, 0x40, 0x40, 0x40, 0x3F}
	case 'V':
		pattern = []uint8{0x1F, 0x20, 0x40, 0x20, 0x1F}
	case 'W':
		pattern = []uint8{0x7F, 0x20, 0x18, 0x20, 0x7F}
	case 'X':
		pattern = []uint8{0x63, 0x14, 0x08, 0x14, 0x63}
	case 'Y':
		pattern = []uint8{0x07, 0x08, 0x70, 0x08, 0x07}
	case 'Z':
		pattern = []uint8{0x61, 0x51, 0x49, 0x45, 0x43}

	// Lowercase letters
	case 'a':
		pattern = []uint8{0x20, 0x54, 0x54, 0x54, 0x78}
	case 'b':
		pattern = []uint8{0x7F, 0x48, 0x44, 0x44, 0x38}
	case 'c':
		pattern = []uint8{0x38, 0x44, 0x44, 0x44, 0x20}
	case 'd':
		pattern = []uint8{0x38, 0x44, 0x44, 0x48, 0x7F}
	case 'e':
		pattern = []uint8{0x38, 0x54, 0x54, 0x54, 0x18}
	case 'f':
		pattern = []uint8{0x08, 0x7E, 0x09, 0x01, 0x02}
	case 'g':
		pattern = []uint8{0x0C, 0x52, 0x52, 0x52, 0x3E}
	case 'h':
		pattern = []uint8{0x7F, 0x08, 0x04, 0x04, 0x78}
	case 'i':
		pattern = []uint8{0x00, 0x44, 0x7D, 0x40, 0x00}
	case 'j':
		pattern = []uint8{0x20, 0x40, 0x44, 0x3D, 0x00}
	case 'k':
		pattern = []uint8{0x7F, 0x10, 0x28, 0x44, 0x00}
	case 'l':
		pattern = []uint8{0x00, 0x41, 0x7F, 0x40, 0x00}
	case 'm':
		pattern = []uint8{0x7C, 0x04, 0x18, 0x04, 0x78}
	case 'n':
		pattern = []uint8{0x7C, 0x08, 0x04, 0x04, 0x78}
	case 'o':
		pattern = []uint8{0x38, 0x44, 0x44, 0x44, 0x38}
	case 'p':
		pattern = []uint8{0x7C, 0x14, 0x14, 0x14, 0x08}
	case 'q':
		pattern = []uint8{0x08, 0x14, 0x14, 0x18, 0x7C}
	case 'r':
		pattern = []uint8{0x7C, 0x08, 0x04, 0x04, 0x08}
	case 's':
		pattern = []uint8{0x48, 0x54, 0x54, 0x54, 0x20}
	case 't':
		pattern = []uint8{0x04, 0x3F, 0x44, 0x40, 0x20}
	case 'u':
		pattern = []uint8{0x3C, 0x40, 0x40, 0x20, 0x7C}
	case 'v':
		pattern = []uint8{0x1C, 0x20, 0x40, 0x20, 0x1C}
	case 'w':
		pattern = []uint8{0x3C, 0x40, 0x30, 0x40, 0x3C}
	case 'x':
		pattern = []uint8{0x44, 0x28, 0x10, 0x28, 0x44}
	case 'y':
		pattern = []uint8{0x0C, 0x50, 0x50, 0x50, 0x3C}
	case 'z':
		pattern = []uint8{0x44, 0x64, 0x54, 0x4C, 0x44}

	// Numbers
	case '0':
		pattern = []uint8{0x3E, 0x51, 0x49, 0x45, 0x3E}
	case '1':
		pattern = []uint8{0x00, 0x42, 0x7F, 0x40, 0x00}
	case '2':
		pattern = []uint8{0x62, 0x51, 0x49, 0x45, 0x42}
	case '3':
		pattern = []uint8{0x42, 0x41, 0x49, 0x49, 0x36}
	case '4':
		pattern = []uint8{0x1C, 0x12, 0x11, 0x7F, 0x10}
	case '5':
		pattern = []uint8{0x47, 0x45, 0x45, 0x45, 0x39}
	case '6':
		pattern = []uint8{0x3C, 0x4A, 0x49, 0x49, 0x30}
	case '7':
		pattern = []uint8{0x01, 0x71, 0x09, 0x05, 0x03}
	case '8':
		pattern = []uint8{0x36, 0x49, 0x49, 0x49, 0x36}
	case '9':
		pattern = []uint8{0x06, 0x49, 0x49, 0x29, 0x1E}

	// Special characters
	case ' ':
		pattern = []uint8{0x00, 0x00, 0x00, 0x00, 0x00}
	case '.':
		pattern = []uint8{0x00, 0x60, 0x60, 0x00, 0x00}
	case ',':
		pattern = []uint8{0x00, 0x80, 0x60, 0x00, 0x00}
	case ':':
		pattern = []uint8{0x00, 0x36, 0x36, 0x00, 0x00}
	case ';':
		pattern = []uint8{0x00, 0x80, 0x36, 0x00, 0x00}
	case '!':
		pattern = []uint8{0x00, 0x00, 0x5F, 0x00, 0x00}
	case '?':
		pattern = []uint8{0x02, 0x01, 0x51, 0x09, 0x06}
	case '-':
		pattern = []uint8{0x08, 0x08, 0x08, 0x08, 0x08}
	case '_':
		pattern = []uint8{0x80, 0x80, 0x80, 0x80, 0x80}
	case '+':
		pattern = []uint8{0x08, 0x08, 0x3E, 0x08, 0x08}
	case '=':
		pattern = []uint8{0x14, 0x14, 0x14, 0x14, 0x14}
	case '/':
		pattern = []uint8{0x20, 0x10, 0x08, 0x04, 0x02}
	case '\\':
		pattern = []uint8{0x02, 0x04, 0x08, 0x10, 0x20}
	case '(':
		pattern = []uint8{0x00, 0x1C, 0x22, 0x41, 0x00}
	case ')':
		pattern = []uint8{0x00, 0x41, 0x22, 0x1C, 0x00}
	case '[':
		pattern = []uint8{0x00, 0x7F, 0x41, 0x41, 0x00}
	case ']':
		pattern = []uint8{0x00, 0x41, 0x41, 0x7F, 0x00}
	case '{':
		pattern = []uint8{0x00, 0x08, 0x36, 0x41, 0x00}
	case '}':
		pattern = []uint8{0x00, 0x41, 0x36, 0x08, 0x00}
	case '>':
		pattern = []uint8{0x08, 0x14, 0x22, 0x41, 0x00}

	default:
		// Unknown character displays as block
		pattern = []uint8{0x7F, 0x41, 0x41, 0x41, 0x7F}
	}

	// Draw character pixels
	for col := 0; col < 5; col++ {
		colData := pattern[col]
		for row := 0; row < 7; row++ {
			if (colData & (1 << row)) != 0 {
				display.SetPixel(x+int16(col), y+int16(row), color.RGBA{255, 255, 255, 255})
			}
		}
	}
}

// Play beep sound - supplementary missing function
func playBeep(frequency int, duration time.Duration) {
	// Simple beep implementation - using speaker pin
	// This is just an example, actual implementation needs PWM or other audio technology
	period := time.Duration(1000000/frequency) * time.Microsecond
	half := period / 2

	start := time.Now()
	for time.Since(start) < duration {
		speakerDOUTPin.High()
		time.Sleep(half)
		speakerDOUTPin.Low()
		time.Sleep(half)
	}
}

// Display boot screen
func showBootScreen() {
	if !oledInitialized {
		return
	}

	display.ClearBuffer()
	drawSimpleText("ESP32", 35, 10)
	drawSimpleText("AI ASSISTANT", 10, 25)
	drawSimpleText("STARTING...", 15, 40)
	display.Display()
}

// Display status information
func showStatus(state string) {
	if !oledInitialized {
		return
	}

	display.ClearBuffer()
	drawSimpleText("STATUS:", 0, 0)
	drawSimpleText(state, 0, 15)
	drawSimpleText("READY", 0, 35)
	display.Display()
}
