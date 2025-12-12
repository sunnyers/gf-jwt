package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fintech-canada/redbook-wechat-monitoring/internal/config"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/database"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/service"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/web"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main [OPTION]",
		Brief: "启动Redbook和WeChat对话监控系统",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 获取配置文件路径
			configPath := parser.GetOpt("config", "./config.yaml").String()
			if configPath == "" {
				// 尝试从环境变量获取
				if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
					configPath = envPath
				} else {
					configPath = "./config.yaml"
				}
			}
			
			// 加载配置
			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}

			// 初始化日志
			logger := glog.New()
			// 设置日志级别 (debug=1, info=2, warn=3, error=4)
			levelMap := map[string]int{
				"debug": 1,
				"info":  2,
				"warn":  3,
				"error": 4,
			}
			if level, ok := levelMap[cfg.Logging.Level]; ok {
				logger.SetLevel(level)
			}
			if cfg.Logging.File != "" {
				logger.SetPath(cfg.Logging.File)
				logger.SetStdoutPrint(false)
			}
			logger.Info(ctx, "配置加载成功")

			// 初始化数据库
			db, err := database.Initialize(ctx, cfg.Database)
			if err != nil {
				return fmt.Errorf("数据库初始化失败: %w", err)
			}
			logger.Info(ctx, "数据库初始化成功")

			// 初始化服务
			svc, err := service.New(ctx, cfg, db)
			if err != nil {
				return fmt.Errorf("服务初始化失败: %w", err)
			}
			logger.Info(ctx, "服务初始化成功")

			// 启动后台任务（Google Sheets同步等）
			if err := svc.StartBackgroundTasks(ctx); err != nil {
				return fmt.Errorf("启动后台任务失败: %w", err)
			}
			logger.Info(ctx, "后台任务启动成功")

			// 启动Web服务器
			s := g.Server()
			s.SetPort(cfg.Server.Port)
			// GoFrame会自动监听配置的端口

			// 注册路由
			web.RegisterRoutes(s, svc, cfg)

			// 启动服务器
			logger.Infof(ctx, "服务器启动在 http://%s:%d", cfg.Server.Host, cfg.Server.Port)
			s.Run()

			return nil
		},
	}
)
