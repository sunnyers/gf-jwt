package model

import (
	"time"

	"gorm.io/gorm"
)

// Conversation 对话记录
type Conversation struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Platform      string         `gorm:"type:varchar(20);not null;index" json:"platform"` // redbook, wechat
	AccountID     string         `gorm:"type:varchar(100);not null;index" json:"account_id"` // 账户ID
	AccountName   string         `gorm:"type:varchar(200)" json:"account_name"` // 账户名称
	CustomerID    string         `gorm:"type:varchar(100);not null;index" json:"customer_id"` // 客户ID
	CustomerName  string         `gorm:"type:varchar(200)" json:"customer_name"` // 客户名称
	CustomerPhone string         `gorm:"type:varchar(50)" json:"customer_phone"` // 客户电话
	Status        string         `gorm:"type:varchar(20);default:'active'" json:"status"` // active, closed
	StartedAt     time.Time      `gorm:"not null" json:"started_at"`
	LastMessageAt time.Time      `gorm:"not null;index" json:"last_message_at"`
	MessageCount  int            `gorm:"default:0" json:"message_count"`
	SyncedToSheet bool           `gorm:"default:false;index" json:"synced_to_sheet"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// Message 消息记录
type Message struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ConversationID uint           `gorm:"not null;index" json:"conversation_id"`
	Platform       string         `gorm:"type:varchar(20);not null" json:"platform"`
	MessageID      string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"message_id"` // 平台消息ID
	SenderID       string         `gorm:"type:varchar(100);not null" json:"sender_id"` // 发送者ID
	SenderName     string         `gorm:"type:varchar(200)" json:"sender_name"` // 发送者名称
	SenderType     string         `gorm:"type:varchar(20);not null" json:"sender_type"` // customer, staff
	Content        string         `gorm:"type:text" json:"content"` // 消息内容
	ContentType    string         `gorm:"type:varchar(50);default:'text'" json:"content_type"` // text, image, video, file, etc.
	MediaURL       string         `gorm:"type:varchar(500)" json:"media_url"` // 媒体文件URL
	Timestamp      time.Time      `gorm:"not null;index" json:"timestamp"`
	SyncedToSheet  bool           `gorm:"default:false;index" json:"synced_to_sheet"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"`
}

// Account 账户信息
type Account struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Platform    string         `gorm:"type:varchar(20);not null;index" json:"platform"`
	AccountID   string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"account_id"`
	AccountName string         `gorm:"type:varchar(200);not null" json:"account_name"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON格式的额外信息
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// SyncLog 同步日志
type SyncLog struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	SyncType      string         `gorm:"type:varchar(50);not null" json:"sync_type"` // google_sheets
	Status        string         `gorm:"type:varchar(20);not null" json:"status"` // success, failed
	RecordsSynced int            `gorm:"default:0" json:"records_synced"`
	ErrorMessage  string         `gorm:"type:text" json:"error_message"`
	StartedAt     time.Time      `gorm:"not null" json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
