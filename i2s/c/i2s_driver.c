#include "i2s_driver.h"
#include "driver/i2s_std.h"
#include "esp_log.h"
#include <stdlib.h>

static const char* TAG = "I2S_DRIVER";
static bool initialized = false;
static i2s_chan_handle_t rx_handle = NULL;

int my_i2s_init(my_i2s_config_t* config) {
    if (initialized) {
        return 0; // 已初始化
    }

    // I2S通道配置
    i2s_chan_config_t chan_cfg = I2S_CHANNEL_DEFAULT_CONFIG(I2S_NUM_AUTO, I2S_ROLE_MASTER);
    chan_cfg.dma_desc_num = config->dma_buf_count;
    chan_cfg.dma_frame_num = config->dma_buf_len;
    
    // 创建I2S RX通道
    esp_err_t ret = i2s_new_channel(&chan_cfg, NULL, &rx_handle);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "I2S new channel failed: %s", esp_err_to_name(ret));
        return -1;
    }

    // I2S标准配置
    i2s_std_config_t std_cfg = {
        .clk_cfg = I2S_STD_CLK_DEFAULT_CONFIG(config->sample_rate),
        .slot_cfg = I2S_STD_PHILIPS_SLOT_DEFAULT_CONFIG(I2S_DATA_BIT_WIDTH_32BIT, I2S_SLOT_MODE_MONO),
        .gpio_cfg = {
            .mclk = I2S_GPIO_UNUSED,
            .bclk = config->sck_pin,
            .ws = config->ws_pin,
            .dout = I2S_GPIO_UNUSED,
            .din = config->din_pin,
            .invert_flags = {
                .mclk_inv = false,
                .bclk_inv = false,
                .ws_inv = false,
            },
        },
    };

    // 初始化I2S标准模式
    ret = i2s_channel_init_std_mode(rx_handle, &std_cfg);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "I2S init std mode failed: %s", esp_err_to_name(ret));
        i2s_del_channel(rx_handle);
        rx_handle = NULL;
        return -1;
    }

    // 启用I2S通道
    ret = i2s_channel_enable(rx_handle);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "I2S enable failed: %s", esp_err_to_name(ret));
        i2s_del_channel(rx_handle);
        rx_handle = NULL;
        return -1;
    }

    initialized = true;
    ESP_LOGI(TAG, "I2S initialized successfully");
    return 0;
}

int my_i2s_read_samples(int16_t* buffer, int samples, int timeout_ms) {
    if (!initialized || !rx_handle) {
        return -1;
    }

    size_t bytes_read = 0;
    int32_t* temp_buffer = malloc(samples * sizeof(int32_t));
    if (!temp_buffer) {
        return -1;
    }

    esp_err_t ret = i2s_channel_read(rx_handle, temp_buffer, 
                                    samples * sizeof(int32_t), 
                                    &bytes_read, 
                                    timeout_ms);
    
    if (ret == ESP_OK) {
        // 转换32位到16位
        int samples_read = bytes_read / sizeof(int32_t);
        for (int i = 0; i < samples_read; i++) {
            buffer[i] = (int16_t)(temp_buffer[i] >> 16);
        }
        free(temp_buffer);
        return samples_read;
    }

    free(temp_buffer);
    return -1;
}

int my_i2s_write_samples(int16_t* buffer, int samples, int timeout_ms) {
    // 暂时不实现写功能
    return -1;
}

int my_i2s_deinit(void) {
    if (!initialized || !rx_handle) {
        return 0;
    }
    
    // 禁用通道
    i2s_channel_disable(rx_handle);
    
    // 删除通道
    esp_err_t ret = i2s_del_channel(rx_handle);
    rx_handle = NULL;
    initialized = false;
    
    return (ret == ESP_OK) ? 0 : -1;
}

bool my_i2s_is_initialized(void) {
    return initialized;
}