#!/bin/bash

# 项目名称
PROJECT_NAME="code-share"

# 目标平台和架构
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# 清理旧的构建目录
rm -rf build
mkdir -p build

# 遍历所有平台和架构
for PLATFORM in "${PLATFORMS[@]}"; do
    # 分割平台和架构
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}

    # 设置输出文件名
    OUTPUT_NAME="${PROJECT_NAME}-${GOOS}-${GOARCH}"

    # 如果是 Windows 平台，添加 .exe 后缀
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    # 编译
    echo "Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -o "build/$OUTPUT_NAME"

    # 压缩
    if [ "$GOOS" == "windows" ]; then
        # 使用 zip 压缩 Windows 平台的可执行文件
        echo "Compressing $OUTPUT_NAME..."
        zip -j "build/${PROJECT_NAME}-${GOOS}-${GOARCH}.zip" "build/$OUTPUT_NAME"
    else
        # 使用 tar 压缩其他平台的可执行文件
        echo "Compressing $OUTPUT_NAME..."
        tar -czf "build/${PROJECT_NAME}-${GOOS}-${GOARCH}.tar.gz" -C "build" "$OUTPUT_NAME"
    fi

    # 删除未压缩的可执行文件
    rm "build/$OUTPUT_NAME"
done

echo "Build and compression completed!"
