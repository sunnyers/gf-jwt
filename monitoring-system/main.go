package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fintech-canada/redbook-wechat-monitoring/internal/cmd"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()
	
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "./config.yaml", "配置文件路径")
	flag.Parse()

	// 设置配置路径到环境变量（供cmd包使用）
	if configPath != "" {
		os.Setenv("CONFIG_PATH", configPath)
	}

	// 启动服务
	cmd.Main.Run(ctx)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n正在关闭服务...")
	os.Exit(0)
}
