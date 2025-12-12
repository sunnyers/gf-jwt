package repository

import (
	"context"
	"time"

	"github.com/fintech-canada/redbook-wechat-monitoring/internal/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetConversations 获取对话列表
func (r *Repository) GetConversations(ctx context.Context, filters map[string]interface{}) ([]interface{}, error) {
	var conversations []model.Conversation
	query := r.db.WithContext(ctx).Preload("Messages")

	// 应用过滤器
	if platform, ok := filters["platform"].(string); ok && platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if accountID, ok := filters["account_id"].(string); ok && accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// 分页
	page := 1
	pageSize := 20
	if p, ok := filters["page"].(int); ok && p > 0 {
		page = p
	}
	if ps, ok := filters["page_size"].(int); ok && ps > 0 {
		pageSize = ps
	}
	offset := (page - 1) * pageSize

	if err := query.Order("last_message_at DESC").Offset(offset).Limit(pageSize).Find(&conversations).Error; err != nil {
		return nil, err
	}

	result := make([]interface{}, len(conversations))
	for i, conv := range conversations {
		result[i] = conv
	}
	return result, nil
}

// GetMessages 获取消息列表
func (r *Repository) GetMessages(ctx context.Context, conversationID uint) ([]interface{}, error) {
	var messages []model.Message
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("timestamp ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}

	result := make([]interface{}, len(messages))
	for i, msg := range messages {
		result[i] = msg
	}
	return result, nil
}

// CreateOrUpdateConversation 创建或更新对话
func (r *Repository) CreateOrUpdateConversation(ctx context.Context, conv *model.Conversation) error {
	var existing model.Conversation
	err := r.db.WithContext(ctx).
		Where("platform = ? AND customer_id = ? AND account_id = ?", conv.Platform, conv.CustomerID, conv.AccountID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新对话
		return r.db.WithContext(ctx).Create(conv).Error
	} else if err != nil {
		return err
	}

	// 更新现有对话
	existing.LastMessageAt = time.Now()
	existing.MessageCount = existing.MessageCount + 1
	if conv.Status != "" {
		existing.Status = conv.Status
	}
	return r.db.WithContext(ctx).Save(&existing).Error
}

// CreateMessage 创建消息
func (r *Repository) CreateMessage(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// GetUnsyncedMessages 获取未同步的消息
func (r *Repository) GetUnsyncedMessages(ctx context.Context, limit int) ([]model.Message, error) {
	var messages []model.Message
	if err := r.db.WithContext(ctx).
		Where("synced_to_sheet = ?", false).
		Order("timestamp ASC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// MarkMessagesAsSynced 标记消息为已同步
func (r *Repository) MarkMessagesAsSynced(ctx context.Context, messageIDs []uint) error {
	return r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("id IN ?", messageIDs).
		Update("synced_to_sheet", true).Error
}
