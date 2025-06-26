package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
)

// 定义引脚
var (
	// LED指示灯
	statusLED = machine.GPIO2

	// 启动按钮 (BOOT BUTTON)
	bootButton = machine.GPIO34 // A6引脚

	// OLED显示屏 I2C引脚
	sclPin = machine.GPIO22 // OLED SCL引脚
	sdaPin = machine.GPIO21 // OLED SDA引脚

	// 麦克风引脚 (INMP441)
	micDINPin = machine.GPIO27 // 麦克风 DIN引脚
	micWSPin  = machine.GPIO26 // 麦克风 WS引脚
	micSCKPin = machine.GPIO25 // 麦克风 SCK引脚

	// 扬声器引脚
	speakerDOUTPin = machine.GPIO23 // 扬声器 DOUT引脚
	speakerBCLKPin = machine.GPIO33 // 扬声器 BCLK引脚
	speakerWSPin   = machine.GPIO32 // 扬声器 WS引脚
)

// 系统状态
type SystemState int

const (
	StateIdle SystemState = iota
	StateRecording
	StatePlaying
	StateProcessing
)

// 音频配置
const (
	SAMPLE_RATE     = 16000 // 16kHz采样率
	BITS_PER_SAMPLE = 16    // 16位采样
	MAX_RECORD_TIME = 10    // 最大录音时间（秒）
	BUFFER_SIZE     = 1024  // 音频缓冲区大小
)

// 全局变量
var (
	currentState     = StateIdle
	recordingTime    = 0.0 // 录音时长（秒）
	playingTime      = 0.0 // 播放时长（秒）
	display          ssd1306.Device
	oledInitialized  = false
	audioBuffer      = make([]int16, SAMPLE_RATE*MAX_RECORD_TIME) // 音频缓冲区
	recordedSamples  = 0                                          // 已录制的样本数
	playbackPosition = 0                                          // 播放位置
	i2sInitialized   = false
)

func main() {
	fmt.Println("ESP32 录音播放系统启动中...")

	// 初始化硬件
	initHardware()

	// 显示启动信息
	displayMessage("ESP32", "录音系统就绪")
	time.Sleep(2 * time.Second)

	// 主循环
	for {
		// 检查按键状态
		checkButtonPress()

		// 根据当前状态执行相应操作
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

// 初始化硬件
func initHardware() {
	fmt.Println("初始化硬件...")

	// 配置LED
	statusLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	statusLED.Low()

	// 配置按键
	bootButton.Configure(machine.PinConfig{Mode: machine.PinInput})

	// 初始化OLED显示屏
	initOLED()

	// 初始化I2S音频接口
	initI2S()

	fmt.Println("硬件初始化完成")
}

// 初始化OLED显示屏
func initOLED() {
	fmt.Println("初始化OLED显示屏...")

	// 配置I2C
	machine.I2C0.Configure(machine.I2CConfig{
		SCL:       sclPin,
		SDA:       sdaPin,
		Frequency: 400000, // 400kHz
	})

	// 创建SSD1306设备实例
	display = ssd1306.NewI2C(machine.I2C0)

	// 配置显示屏
	display.Configure(ssd1306.Config{
		Address: ssd1306.Address_128_32, // 0x3C
		Width:   128,
		Height:  64,
	})

	// 清空显示
	display.ClearDisplay()
	oledInitialized = true
	fmt.Println("OLED显示屏初始化完成")
}

// 初始化I2S音频接口
func initI2S() {
	fmt.Println("初始化I2S音频接口...")

	// 配置麦克风引脚
	micDINPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	micWSPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	micSCKPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// 配置扬声器引脚
	speakerDOUTPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	speakerBCLKPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	speakerWSPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// 初始化I2S配置（这里是简化版本，实际需要根据具体硬件配置）
	i2sInitialized = true
	fmt.Println("I2S音频接口初始化完成")
}

// 检查按键按下
func checkButtonPress() {
	if !bootButton.Get() {
		time.Sleep(50 * time.Millisecond) // 消抖
		if !bootButton.Get() {
			handleButtonPress()
			// 等待按键释放
			for !bootButton.Get() {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

// 处理按键按下事件
func handleButtonPress() {
	fmt.Println("检测到BOOT按键按下")

	// 播放按键确认音
	playBeep(1000, 100*time.Millisecond)

	switch currentState {
	case StateIdle:
		// 开始录音
		startRecording()

	case StateRecording:
		// 停止录音
		stopRecording()

	case StatePlaying:
		// 停止播放
		stopPlaying()

	case StateProcessing:
		// 取消处理
		currentState = StateIdle
		displayMessage("系统", "已取消")
		statusLED.Low()
	}
}

// 开始录音
func startRecording() {
	if !i2sInitialized {
		return
	}

	currentState = StateRecording
	recordingTime = 0.0
	recordedSamples = 0
	statusLED.High()

	displayRecordingStatus()
	fmt.Println("开始录音...")
}

// 停止录音
func stopRecording() {
	if currentState != StateRecording {
		return
	}

	currentState = StateProcessing
	statusLED.Low()

	fmt.Printf("录音完成，时长: %.1f秒\n", recordingTime)
	displayMessage("录音完成", fmt.Sprintf("%.1fs 按键播放", recordingTime))

	// 短暂处理后进入待机状态
	time.Sleep(1 * time.Second)
	currentState = StateIdle
}

// 开始播放
func startPlaying() {
	if recordedSamples == 0 {
		displayMessage("错误", "无录音文件")
		return
	}

	currentState = StatePlaying
	playingTime = 0.0
	playbackPosition = 0
	statusLED.High()

	displayPlayingStatus()
	fmt.Println("开始播放录音...")
}

// 停止播放
func stopPlaying() {
	if currentState != StatePlaying {
		return
	}

	currentState = StateIdle
	statusLED.Low()

	fmt.Println("播放停止")
	displayMessage("播放停止", "按键重新播放")
}

// 处理待机状态
func handleIdleState() {
	// LED慢闪表示待机
	statusLED.High()
	time.Sleep(100 * time.Millisecond)
	statusLED.Low()
	time.Sleep(1900 * time.Millisecond)

	// 显示待机信息
	if recordedSamples > 0 {
		displayMessage("就绪", "按键播放录音")
	} else {
		displayMessage("就绪", "按键开始录音")
	}
}

// 处理录音状态
func handleRecordingState() {
	// LED快闪表示录音中
	statusLED.High()
	time.Sleep(200 * time.Millisecond)
	statusLED.Low()
	time.Sleep(200 * time.Millisecond)

	// 更新录音时间
	recordingTime += 0.4 // 每400ms更新一次

	// 模拟录音数据采集
	if i2sInitialized {
		// 这里应该从I2S接口读取实际的音频数据
		// 现在用模拟数据代替
		samplesThisCycle := int(SAMPLE_RATE * 0.4) // 400ms的样本数
		for i := 0; i < samplesThisCycle && recordedSamples < len(audioBuffer); i++ {
			// 模拟音频数据（实际应该从麦克风读取）
			audioBuffer[recordedSamples] = int16((recordedSamples % 1000) - 500)
			recordedSamples++
		}
	}

	// 更新显示
	displayRecordingStatus()

	// 检查是否达到最大录音时间
	if recordingTime >= MAX_RECORD_TIME {
		stopRecording()
	}
}

// 处理播放状态
func handlePlayingState() {
	// LED常亮表示播放中
	statusLED.High()

	// 更新播放时间
	playingTime += 0.1 // 每100ms更新一次

	// 模拟播放数据输出
	if i2sInitialized {
		// 这里应该向I2S接口输出音频数据
		// 现在用模拟播放代替
		samplesThisCycle := int(SAMPLE_RATE * 0.1) // 100ms的样本数
		playbackPosition += samplesThisCycle
	}

	// 更新显示
	displayPlayingStatus()

	// 检查是否播放完成
	totalPlayTime := float64(recordedSamples) / SAMPLE_RATE
	if playingTime >= totalPlayTime {
		stopPlaying()
	}
}

// 处理处理状态
func handleProcessingState() {
	// LED闪烁表示处理中
	for i := 0; i < 3; i++ {
		statusLED.High()
		time.Sleep(100 * time.Millisecond)
		statusLED.Low()
		time.Sleep(100 * time.Millisecond)
	}
}

// 显示录音状态
func displayRecordingStatus() {
	if !oledInitialized {
		return
	}

	display.ClearBuffer()
	drawSimpleText("录音中...", 0, 0)
	drawSimpleText(fmt.Sprintf("时长: %.1fs", recordingTime), 0, 20)
	drawSimpleText(fmt.Sprintf("最大: %ds", MAX_RECORD_TIME), 0, 40)
	display.Display()
}

// 显示播放状态
func displayPlayingStatus() {
	if !oledInitialized {
		return
	}

	totalTime := float64(recordedSamples) / SAMPLE_RATE
	display.ClearBuffer()
	drawSimpleText("播放中...", 0, 0)
	drawSimpleText(fmt.Sprintf("%.1f/%.1fs", playingTime, totalTime), 0, 20)

	// 显示进度条
	progress := int(playingTime / totalTime * 120)
	if progress > 120 {
		progress = 120
	}
	for i := 0; i < progress; i++ {
		display.SetPixel(int16(i+4), 45, color.RGBA{255, 255, 255, 255})
	}

	display.Display()
}

// 显示消息到屏幕
func displayMessage(title, message string) {
	if !oledInitialized {
		return
	}

	display.ClearBuffer()
	drawSimpleText(title, 0, 0)
	drawSimpleText(message, 0, 20)
	display.Display()

	fmt.Printf("[显示] %s: %s\n", title, message)
}

// 绘制简单文本（使用像素点阵）
func drawSimpleText(text string, x, y int16) {
	charWidth := int16(6) // 包含间距
	currentX := x

	for _, char := range text {
		if currentX > 122 { // 防止超出屏幕边界
			break
		}
		drawChar5x7(char, currentX, y)
		currentX += charWidth
	}
}

// 绘制5x7像素字符 - 优化后的完整字符集
func drawChar5x7(char rune, x, y int16) {
	var pattern []uint8

	switch char {
	// 大写字母
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

	// 小写字母
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

	// 数字
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

	// 特殊字符
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

	default:
		// 未知字符显示为方块
		pattern = []uint8{0x7F, 0x41, 0x41, 0x41, 0x7F}
	}

	// 绘制字符像素
	for col := 0; col < 5; col++ {
		colData := pattern[col]
		for row := 0; row < 7; row++ {
			if (colData & (1 << row)) != 0 {
				display.SetPixel(x+int16(col), y+int16(row), color.RGBA{255, 255, 255, 255})
			}
		}
	}
}

// 播放蜂鸣音 - 补充缺失的函数
func playBeep(frequency int, duration time.Duration) {
	// 简单的蜂鸣实现 - 使用扬声器引脚
	// 这里只是一个示例，实际实现需要PWM或其他音频技术
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

// 显示启动画面
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

// 显示状态信息
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
