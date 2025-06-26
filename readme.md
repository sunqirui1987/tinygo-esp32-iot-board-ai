# TinyGo ESP32 IoT 智能语音板

基于 ESP32 和 TinyGo 开发的智能语音交互板，支持语音录制、播放和 OLED 显示功能。

## 项目特性

- 🎤 支持 I2S 麦克风语音录制（INMP441、MSM261S403）
- 🔊 支持 I2S 功放喇叭音频播放
- 📺 SSD1306 OLED 显示屏界面显示
- 🔘 外接按键交互控制
- ⚡ 基于 TinyGo 轻量级开发

## 硬件外设清单

- **主控芯片**: ESP32
- **音频输入**: I2S麦克风（INMP441、MSM261S403）
- **音频输出**: I2S功放喇叭
- **显示设备**: SSD1306 OLED显示屏
- **交互设备**: 外接按键

## 硬件接口映射表

| ESP32 引脚编号 | 连接外设 | 功能描述 |
|---------------|----------|----------|
| 25 | 麦克风（MIC）| SCK (BCK、BCLK)引脚 |
| 26 | 麦克风（MIC）| WS引脚 |
| 27 | 麦克风（MIC）| DIN (DOUT、DI、DO、DATA、SD) 引脚 |
| 23 | 扬声器（SPEAKER）| DOUT（DIN、DI、DO、DATA） 引脚 |
| 33 | 扬声器（SPEAKER）| BCLK 引脚 |
| 32 | 扬声器（SPEAKER）| WS（LRCK） 引脚 |
| 21 | OLED 显示屏 | SDA 引脚 |
| 22 | OLED 显示屏 | SCL 引脚 |
| 34（A6）| 启动按钮 | BOOT BUTTON |

## 技术规格

- **采样率**: 16kHz
- **采样位数**: 16位
- **最大录音时间**: 10秒
- **音频缓冲区**: 1024字节
- **通信协议**: I2C（OLED）、I2S（音频）

## 快速开始

### 环境要求

- Go 1.22.6+
- TinyGo
- ESP32 开发板

### 安装依赖

```bash
go mod tidy
sh flash.sh
```


