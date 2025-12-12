package wechat

import (
	"crypto/sha1"
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
	config     config.WeChatConfig
	httpClient *http.Client
	accessToken string
	tokenExpiresAt time.Time
}

// Message WeChat消息结构
type Message struct {
	MsgID      string    `json:"MsgId"`
	FromUserID string    `json:"FromUserName"`
	ToUserID   string    `json:"ToUserName"`
	Content    string    `json:"Content"`
	MsgType    string    `json:"MsgType"` // text, image, video, etc.
	CreateTime int64     `json:"CreateTime"`
	CreatedAt  time.Time `json:"created_at"`
}

// WebhookPayload WeChat Webhook请求负载
type WebhookPayload struct {
	ToUserName   string `json:"ToUserName" xml:"ToUserName"`
	FromUserName string `json:"FromUserName" xml:"FromUserName"`
	CreateTime   int64  `json:"CreateTime" xml:"CreateTime"`
	MsgType      string `json:"MsgType" xml:"MsgType"`
	Content      string `json:"Content" xml:"Content"`
	MsgId        string `json:"MsgId" xml:"MsgId"`
}

// NewClient 创建新的WeChat客户端
func NewClient(cfg config.WeChatConfig) (*Client, error) {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// VerifyWebhook 验证Webhook签名（企业微信）
func (c *Client) VerifyWebhook(signature string, timestamp string, nonce string, echostr string) (string, bool) {
	// 企业微信Webhook验证逻辑
	token := c.config.WeCom.WebhookToken
	params := []string{token, timestamp, nonce}
	sort.Strings(params)
	combined := strings.Join(params, "")

	// SHA1加密
	hash := sha1.Sum([]byte(combined))
	expectedSignature := hex.EncodeToString(hash[:])

	if signature == expectedSignature {
		return echostr, true
	}
	return "", false
}

// ParseWebhookMessage 解析Webhook消息
func (c *Client) ParseWebhookMessage(body []byte) (*Message, error) {
	// 企业微信消息解析
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		// 尝试XML解析
		// 这里简化处理，实际应该支持XML格式
		return nil, fmt.Errorf("解析Webhook负载失败: %w", err)
	}

	msg := &Message{
		MsgID:      payload.MsgId,
		FromUserID: payload.FromUserName,
		ToUserID:   payload.ToUserName,
		Content:    payload.Content,
		MsgType:    payload.MsgType,
		CreateTime: payload.CreateTime,
		CreatedAt:  time.Unix(payload.CreateTime, 0),
	}

	return msg, nil
}

// GetAccessToken 获取企业微信访问令牌
func (c *Client) GetAccessToken() (string, error) {
	// 检查token是否过期
	if c.accessToken != "" && time.Now().Before(c.tokenExpiresAt) {
		return c.accessToken, nil
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		c.config.WeCom.CorpID, c.config.WeCom.CorpSecret)

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
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("获取访问令牌失败: %s", result.ErrMsg)
	}

	c.accessToken = result.AccessToken
	c.tokenExpiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return c.accessToken, nil
}
