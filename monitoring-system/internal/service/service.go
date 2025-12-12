package service

import (
	"context"
	"fmt"
	"time"

	"github.com/fintech-canada/redbook-wechat-monitoring/internal/config"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/integration/google_sheets"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/integration/redbook"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/integration/wechat"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/model"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/repository"
	"gorm.io/gorm"
)

type Service struct {
	cfg        *config.Config
	db         *gorm.DB
	repo       *repository.Repository
	redbook    *redbook.Client
	wechat     *wechat.Client
	sheets     *google_sheets.Client
}

// New 创建新的服务实例
func New(ctx context.Context, cfg *config.Config, db *gorm.DB) (*Service, error) {
	repo := repository.New(db)

	var redbookClient *redbook.Client
	var wechatClient *wechat.Client
	var sheetsClient *google_sheets.Client
	var err error

	// 初始化Redbook客户端
	if cfg.Redbook.Enabled {
		redbookClient, err = redbook.NewClient(cfg.Redbook)
		if err != nil {
			return nil, fmt.Errorf("初始化Redbook客户端失败: %w", err)
		}
	}

	// 初始化WeChat客户端
	if cfg.WeChat.Enabled {
		wechatClient, err = wechat.NewClient(cfg.WeChat)
		if err != nil {
			return nil, fmt.Errorf("初始化WeChat客户端失败: %w", err)
		}
	}

	// 初始化Google Sheets客户端
	if cfg.GoogleSheets.Enabled {
		sheetsClient, err = google_sheets.NewClient(ctx, cfg.GoogleSheets)
		if err != nil {
			return nil, fmt.Errorf("初始化Google Sheets客户端失败: %w", err)
		}
	}

	return &Service{
		cfg:     cfg,
		db:      db,
		repo:    repo,
		redbook: redbookClient,
		wechat:  wechatClient,
		sheets:  sheetsClient,
	}, nil
}

// StartBackgroundTasks 启动后台任务
func (s *Service) StartBackgroundTasks(ctx context.Context) error {
	// 启动Google Sheets同步任务
	if s.sheets != nil && s.cfg.GoogleSheets.Enabled {
		go s.startSheetsSyncTask(ctx)
	}

	return nil
}

// startSheetsSyncTask 启动Google Sheets同步任务
func (s *Service) startSheetsSyncTask(ctx context.Context) {
	if s.sheets == nil {
		return
	}
	s.sheets.SetRepository(s.repo)
	go s.sheets.StartSyncTask(ctx, s.repo)
}

// GetConversations 获取对话列表
func (s *Service) GetConversations(ctx context.Context, filters map[string]interface{}) ([]interface{}, error) {
	return s.repo.GetConversations(ctx, filters)
}

// GetMessages 获取消息列表
func (s *Service) GetMessages(ctx context.Context, conversationID uint) ([]interface{}, error) {
	return s.repo.GetMessages(ctx, conversationID)
}

// ProcessRedbookWebhook 处理Redbook Webhook
func (s *Service) ProcessRedbookWebhook(ctx context.Context, body []byte) error {
	if s.redbook == nil {
		return fmt.Errorf("Redbook客户端未初始化")
	}

	msg, err := s.redbook.ParseWebhookMessage(body)
	if err != nil {
		return fmt.Errorf("解析Redbook消息失败: %w", err)
	}

	return s.ProcessRedbookMessage(ctx, msg)
}

// ProcessRedbookMessage 处理Redbook消息
func (s *Service) ProcessRedbookMessage(ctx context.Context, msg *redbook.Message) error {
	// 查找或创建对话
	conv := &model.Conversation{
		Platform:      "redbook",
		AccountID:     msg.ToUserID,
		CustomerID:    msg.FromUserID,
		Status:        "active",
		StartedAt:     msg.CreatedAt,
		LastMessageAt: msg.CreatedAt,
		MessageCount:  1,
	}

	if err := s.repo.CreateOrUpdateConversation(ctx, conv); err != nil {
		return fmt.Errorf("保存对话失败: %w", err)
	}

	// 保存消息
	message := &model.Message{
		ConversationID: conv.ID,
		Platform:       "redbook",
		MessageID:      msg.MsgID,
		SenderID:       msg.FromUserID,
		SenderType:     "customer",
		Content:        msg.Content,
		ContentType:    msg.MsgType,
		Timestamp:      msg.CreatedAt,
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return fmt.Errorf("保存消息失败: %w", err)
	}

	return nil
}

// ProcessWeChatWebhook 处理WeChat Webhook
func (s *Service) ProcessWeChatWebhook(ctx context.Context, body []byte) error {
	if s.wechat == nil {
		return fmt.Errorf("WeChat客户端未初始化")
	}

	msg, err := s.wechat.ParseWebhookMessage(body)
	if err != nil {
		return fmt.Errorf("解析WeChat消息失败: %w", err)
	}

	return s.ProcessWeChatMessage(ctx, msg)
}

// ProcessWeChatMessage 处理WeChat消息
func (s *Service) ProcessWeChatMessage(ctx context.Context, msg *wechat.Message) error {
	// 查找或创建对话
	conv := &model.Conversation{
		Platform:      "wechat",
		AccountID:     msg.ToUserID,
		CustomerID:    msg.FromUserID,
		Status:        "active",
		StartedAt:     msg.CreatedAt,
		LastMessageAt: msg.CreatedAt,
		MessageCount:  1,
	}

	if err := s.repo.CreateOrUpdateConversation(ctx, conv); err != nil {
		return fmt.Errorf("保存对话失败: %w", err)
	}

	// 保存消息
	message := &model.Message{
		ConversationID: conv.ID,
		Platform:       "wechat",
		MessageID:      msg.MsgID,
		SenderID:       msg.FromUserID,
		SenderType:     "customer", // 需要根据实际情况判断
		Content:        msg.Content,
		ContentType:    msg.MsgType,
		Timestamp:      msg.CreatedAt,
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return fmt.Errorf("保存消息失败: %w", err)
	}

	return nil
}

// VerifyRedbookWebhook 验证Redbook Webhook签名
func (s *Service) VerifyRedbookWebhook(signature, timestamp, nonce string, body []byte) bool {
	if s.redbook == nil {
		return false
	}
	return s.redbook.VerifyWebhook(signature, timestamp, nonce, body)
}

// VerifyWeChatWebhook 验证WeChat Webhook签名
func (s *Service) VerifyWeChatWebhook(signature, timestamp, nonce, echostr string) (string, bool) {
	if s.wechat == nil {
		return "", false
	}
	return s.wechat.VerifyWebhook(signature, timestamp, nonce, echostr)
}

// GetStats 获取统计信息
func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// TODO: 实现统计信息查询
	return map[string]interface{}{
		"total_conversations": 0,
		"total_messages":      0,
		"today_conversations": 0,
		"today_messages":      0,
	}, nil
}

// TriggerSync 手动触发同步
func (s *Service) TriggerSync(ctx context.Context) error {
	if s.sheets == nil {
		return fmt.Errorf("Google Sheets客户端未初始化")
	}

	// 获取未同步的消息
	messages, err := s.repo.GetUnsyncedMessages(ctx, 1000)
	if err != nil {
		return fmt.Errorf("获取未同步消息失败: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	// 同步到Google Sheets
	if err := s.sheets.SyncMessages(ctx, messages); err != nil {
		return fmt.Errorf("同步失败: %w", err)
	}

	// 标记为已同步
	messageIDs := make([]uint, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.ID
	}

	if err := s.repo.MarkMessagesAsSynced(ctx, messageIDs); err != nil {
		return fmt.Errorf("标记消息失败: %w", err)
	}

	return nil
}
