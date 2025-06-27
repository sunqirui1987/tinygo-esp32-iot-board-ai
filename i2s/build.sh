#!/bin/bash

# 设置ESP-IDF环境
source $IDF_PATH/export.sh

# 创建临时项目目录
mkdir -p temp_project/main
mkdir -p temp_project/components/i2s_driver

# 复制源文件
cp c/* temp_project/components/i2s_driver/

# 创建主程序（用于编译）
cat > temp_project/main/main.c << 'EOF'
#include "i2s_driver.h"
void app_main(void) {
    // 空主程序，仅用于编译
}
EOF

# 创建主CMakeLists.txt
cat > temp_project/CMakeLists.txt << 'EOF'
cmake_minimum_required(VERSION 3.16)
include($ENV{IDF_PATH}/tools/cmake/project.cmake)
project(i2s_lib)
EOF

# 创建main的CMakeLists.txt
cat > temp_project/main/CMakeLists.txt << 'EOF'
idf_component_register(
    SRCS "main.c"
    REQUIRES i2s_driver
)
EOF

# 编译项目
cd temp_project
idf.py build

# 创建lib目录（如果不存在）
mkdir -p ../lib

# 提取静态库
cp build/esp-idf/i2s_driver/libi2s_driver.a ../lib/libi2s.a

# 清理
cd ..
rm -rf temp_project

echo "I2S静态库编译完成: lib/libi2s.a"