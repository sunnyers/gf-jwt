package redbook

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/fintech-canada/redbook-wechat-monitoring/internal/config"
)

type Client struct {
	config     config.RedbookConfig
	httpClient *http.Client
}

// Message Redbook消息结构
type Message struct {
	MsgID      string    `json:"msg_id"`
	FromUserID string    `json:"from_user_id"`
	ToUserID   string    `json:"to_user_id"`
	Content    string    `json:"content"`
	MsgType    string    `json:"msg_type"` // text, image, video, etc.
	Timestamp  int64     `json:"timestamp"`
	CreatedAt  time.Time `json:"created_at"`
}

// WebhookPayload Webhook请求负载
type WebhookPayload struct {
	EventType string          `json:"event_type"` // message, user_follow, etc.
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// NewClient 创建新的Redbook客户端
func NewClient(cfg config.RedbookConfig) (*Client, error) {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// VerifyWebhook 验证Webhook签名
func (c *Client) VerifyWebhook(signature string, timestamp string, nonce string, body []byte) bool {
	// 小红书Webhook验证逻辑
	// 将token、timestamp、nonce按字典序排序后拼接
	params := []string{c.config.WebhookToken, timestamp, nonce}
	sort.Strings(params)
	combined := strings.Join(params, "")

	// SHA256加密
	hash := sha256.Sum256([]byte(combined))
	expectedSignature := hex.EncodeToString(hash[:])

	return signature == expectedSignature
}

// ParseWebhookMessage 解析Webhook消息
func (c *Client) ParseWebhookMessage(body []byte) (*Message, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("解析Webhook负载失败: %w", err)
	}

	if payload.EventType != "message" {
		return nil, fmt.Errorf("不支持的事件类型: %s", payload.EventType)
	}

	var msg Message
	if err := json.Unmarshal(payload.Data, &msg); err != nil {
		return nil, fmt.Errorf("解析消息数据失败: %w", err)
	}

	msg.CreatedAt = time.Unix(payload.Timestamp, 0)
	return &msg, nil
}

// GetAccessToken 获取访问令牌
func (c *Client) GetAccessToken() (string, error) {
	// 实现获取访问令牌的逻辑
	// 小红书API通常需要app_id和app_secret来获取token
	url := fmt.Sprintf("https://api.xiaohongshu.com/oauth/access_token?app_id=%s&app_secret=%s&grant_type=client_credentials",
		c.config.AppID, c.config.AppSecret)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("请求访问令牌失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	return result.AccessToken, nil
}
