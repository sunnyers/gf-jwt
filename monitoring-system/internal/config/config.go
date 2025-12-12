package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Redbook     RedbookConfig     `yaml:"redbook"`
	WeChat      WeChatConfig      `yaml:"wechat"`
	GoogleSheets GoogleSheetsConfig `yaml:"google_sheets"`
	Logging     LoggingConfig     `yaml:"logging"`
	Security    SecurityConfig    `yaml:"security"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Type   string            `yaml:"type"`
	SQLite SQLiteConfig      `yaml:"sqlite"`
	MySQL  MySQLConfig       `yaml:"mysql"`
	Postgres PostgresConfig  `yaml:"postgres"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type RedbookConfig struct {
	Enabled      bool   `yaml:"enabled"`
	AppID        string `yaml:"app_id"`
	AppSecret    string `yaml:"app_secret"`
	WebhookToken string `yaml:"webhook_token"`
	WebhookURL   string `yaml:"webhook_url"`
}

type WeChatConfig struct {
	Enabled      bool          `yaml:"enabled"`
	WeCom        WeComConfig   `yaml:"wecom"`
	Normal       NormalWeChatConfig `yaml:"normal"`
}

type WeComConfig struct {
	CorpID      string `yaml:"corp_id"`
	CorpSecret  string `yaml:"corp_secret"`
	AgentID     string `yaml:"agent_id"`
	WebhookToken string `yaml:"webhook_token"`
	WebhookURL  string `yaml:"webhook_url"`
}

type NormalWeChatConfig struct {
	AppID        string `yaml:"app_id"`
	AppSecret    string `yaml:"app_secret"`
	WebhookToken string `yaml:"webhook_token"`
}

type GoogleSheetsConfig struct {
	Enabled        bool   `yaml:"enabled"`
	SpreadsheetID  string `yaml:"spreadsheet_id"`
	CredentialsFile string `yaml:"credentials_file"`
	SheetName      string `yaml:"sheet_name"`
	SyncInterval   int    `yaml:"sync_interval"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

type SecurityConfig struct {
	JWTSecret    string `yaml:"jwt_secret"`
	AdminUsername string `yaml:"admin_username"`
	AdminPassword string `yaml:"admin_password"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证必需的配置项
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port == 0 {
		return fmt.Errorf("服务器端口未配置")
	}

	if c.Database.Type == "" {
		return fmt.Errorf("数据库类型未配置")
	}

	if c.Redbook.Enabled && c.Redbook.AppID == "" {
		return fmt.Errorf("Redbook AppID未配置")
	}

	if c.WeChat.Enabled && c.WeChat.WeCom.CorpID == "" {
		return fmt.Errorf("WeChat/WeCom CorpID未配置")
	}

	if c.GoogleSheets.Enabled && c.GoogleSheets.SpreadsheetID == "" {
		return fmt.Errorf("Google Sheets SpreadsheetID未配置")
	}

	return nil
}
