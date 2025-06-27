#ifndef I2S_DRIVER_H
#define I2S_DRIVER_H

#include <stdint.h>
#include <stdbool.h>

// I2S配置结构（重命名避免冲突）
typedef struct {
    int sample_rate;
    int bits_per_sample;
    int dma_buf_count;
    int dma_buf_len;
    int ws_pin;
    int sck_pin;
    int din_pin;
} my_i2s_config_t;

// 导出函数（重命名避免冲突）
int my_i2s_init(my_i2s_config_t* config);
int my_i2s_deinit(void);
int my_i2s_read_samples(int16_t* buffer, int samples, int timeout_ms);
int my_i2s_write_samples(int16_t* buffer, int samples, int timeout_ms);
bool my_i2s_is_initialized(void);

#endif