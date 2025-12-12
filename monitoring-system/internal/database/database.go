package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fintech-canada/redbook-wechat-monitoring/internal/config"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Initialize 初始化数据库
func Initialize(ctx context.Context, cfg config.DatabaseConfig) (*gorm.DB, error) {
	var err error
	var dialector gorm.Dialector

	switch cfg.Type {
	case "sqlite":
		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(cfg.SQLite.Path), 0755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %w", err)
		}
		dialector = sqlite.Open(cfg.SQLite.Path)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Postgres.Host, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.Port)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.Type)
	}

	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&model.Conversation{},
		&model.Message{},
		&model.Account{},
		&model.SyncLog{},
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return db, nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}
