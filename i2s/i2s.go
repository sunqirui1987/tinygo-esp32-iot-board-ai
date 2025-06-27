package i2s

/*
#cgo CFLAGS: -I./c
#cgo LDFLAGS: -L./lib -li2s
#include "i2s_driver.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// I2S配置
type Config struct {
	SampleRate    int
	BitsPerSample int
	DMABufCount   int
	DMABufLen     int
	WSPin         int
	SCKPin        int
	DINPin        int
}

// 初始化I2S
func Init(config Config) error {
	cConfig := C.my_i2s_config_t{
		sample_rate:     C.int(config.SampleRate),
		bits_per_sample: C.int(config.BitsPerSample),
		dma_buf_count:   C.int(config.DMABufCount),
		dma_buf_len:     C.int(config.DMABufLen),
		ws_pin:          C.int(config.WSPin),
		sck_pin:         C.int(config.SCKPin),
		din_pin:         C.int(config.DINPin),
	}

	ret := C.my_i2s_init(&cConfig)
	if ret != 0 {
		return fmt.Errorf("I2S initialization failed: %d", ret)
	}
	return nil
}

// 读取音频数据
func Read(buffer []int16, timeoutMs int) (int, error) {
	if len(buffer) == 0 {
		return 0, fmt.Errorf("buffer is empty")
	}

	ret := C.my_i2s_read_samples(
		(*C.int16_t)(unsafe.Pointer(&buffer[0])),
		C.int(len(buffer)),
		C.int(timeoutMs),
	)

	if ret < 0 {
		return 0, fmt.Errorf("I2S read failed: %d", ret)
	}

	return int(ret), nil
}

// 检查是否已初始化
func IsInitialized() bool {
	return bool(C.my_i2s_is_initialized())
}

// 反初始化
func Deinit() error {
	ret := C.my_i2s_deinit()
	if ret != 0 {
		return fmt.Errorf("I2S deinitialization failed: %d", ret)
	}
	return nil
}