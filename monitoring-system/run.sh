#!/bin/bash

# Redbook和WeChat对话监控系统启动脚本

set -e

echo "=========================================="
echo "Redbook和WeChat对话监控系统"
echo "=========================================="

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go，请先安装Go 1.21或更高版本"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "错误: Go版本需要1.21或更高，当前版本: $GO_VERSION"
    exit 1
fi

echo "✓ Go版本检查通过: $GO_VERSION"

# 创建必要的目录
echo "创建必要的目录..."
mkdir -p data logs config

# 检查配置文件
CONFIG_FILE="config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "警告: 配置文件 $CONFIG_FILE 不存在"
    if [ -f "config.example.yaml" ]; then
        echo "从示例文件创建配置文件..."
        cp config.example.yaml "$CONFIG_FILE"
        echo "请编辑 $CONFIG_FILE 并填入实际配置"
        exit 1
    else
        echo "错误: 找不到配置文件示例"
        exit 1
    fi
fi

echo "✓ 配置文件检查通过"

# 下载依赖
echo "下载Go依赖..."
go mod download
go mod tidy

echo "✓ 依赖下载完成"

# 构建项目
echo "构建项目..."
go build -o monitoring-system main.go

if [ $? -ne 0 ]; then
    echo "错误: 构建失败"
    exit 1
fi

echo "✓ 构建成功"

# 运行项目
echo ""
echo "启动服务..."
echo "配置文件: $CONFIG_FILE"
echo "访问地址: http://localhost:8080"
echo "健康检查: http://localhost:8080/api/health"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

./monitoring-system -config "$CONFIG_FILE"
