idf_component_register(
    SRCS "i2s_driver.c"
    INCLUDE_DIRS "."
    REQUIRES driver esp_common esp_driver_i2s
)