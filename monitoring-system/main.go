package main

import (
	"context"
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

	// 启动服务
	if err := cmd.Main.Run(ctx, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n正在关闭服务...")
	os.Exit(0)
}
