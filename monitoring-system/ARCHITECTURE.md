# 系统架构和实现原理

## 整体架构

```
┌─────────────────┐         ┌─────────────────┐
│   小红书平台      │         │   企业微信平台    │
│  (Redbook)      │         │   (WeChat/WeCom) │
└────────┬────────┘         └────────┬────────┘
         │                            │
         │ Webhook推送                │ Webhook推送
         │                            │
         ▼                            ▼
┌─────────────────────────────────────────────┐
│          Webhook接收器 (HTTP Server)        │
│  - 验证签名                                 │
│  - 解析消息                                 │
│  - 路由到处理服务                           │
└───────────────┬─────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────┐
│           消息处理服务层                     │
│  - 创建/更新对话记录                        │
│  - 保存消息到数据库                          │
│  - 触发同步任务                              │
└───────────────┬─────────────────────────────┘
                │
                ├─────────────────┐
                │                 │
                ▼                 ▼
    ┌──────────────────┐  ┌──────────────────┐
    │   数据库存储      │  │  Google Sheets    │
    │  (SQLite/MySQL)  │  │   自动同步        │
    └──────────────────┘  └──────────────────┘
```

## 数据流程

### 1. 消息接收流程

```
平台发送消息 
    ↓
Webhook HTTP请求到达服务器
    ↓
验证签名（确保消息来源合法）
    ↓
解析消息内容（JSON/XML）
    ↓
查找或创建对话记录
    ↓
保存消息到数据库
    ↓
标记为未同步状态
    ↓
返回成功响应给平台
```

### 2. 同步流程

```
后台定时任务（每60秒）
    ↓
查询未同步的消息
    ↓
批量准备数据
    ↓
调用Google Sheets API写入
    ↓
标记消息为已同步
    ↓
记录同步日志
```

## 小红书（Redbook）数据获取

### 官方API情况

**小红书开放平台**: https://open.xiaohongshu.com/

小红书提供了**官方开放平台API**，但需要注意：

1. **商业账户API**: 
   - 需要企业认证
   - 需要申请开发者账号
   - 提供消息推送能力（Webhook）

2. **API限制**:
   - 需要审核通过
   - 有调用频率限制
   - 需要付费（根据使用量）

### 实现方式

#### 方式1: Webhook推送（推荐）

小红书平台会主动推送消息到你的服务器：

```go
// Webhook接收端点
POST /api/webhook/redbook

// 请求头
X-Signature: <签名>
X-Timestamp: <时间戳>
X-Nonce: <随机数>

// 请求体（JSON）
{
  "event_type": "message",
  "timestamp": 1234567890,
  "data": {
    "msg_id": "消息ID",
    "from_user_id": "发送者ID",
    "to_user_id": "接收者ID",
    "content": "消息内容",
    "msg_type": "text",
    "timestamp": 1234567890
  }
}
```

**签名验证流程**:
```go
1. 将token、timestamp、nonce按字典序排序
2. 拼接成字符串
3. SHA256加密
4. 与请求头中的signature对比
```

#### 方式2: API轮询（不推荐）

如果Webhook不可用，可以定期调用API获取消息：

```go
// 获取消息列表API
GET https://api.xiaohongshu.com/v1/messages?access_token=xxx

// 需要：
// 1. 获取access_token（使用app_id和app_secret）
// 2. 定期轮询（效率低，不实时）
```

### 配置步骤

1. **注册开发者账号**
   - 访问 https://open.xiaohongshu.com/
   - 完成企业认证
   - 创建应用

2. **获取凭证**
   - App ID
   - App Secret
   - Webhook Token（用于验证）

3. **配置Webhook**
   - 在开放平台设置Webhook URL
   - 格式: `https://your-domain.com/api/webhook/redbook`
   - 必须使用HTTPS

4. **验证连接**
   - 平台会发送验证请求
   - 需要正确响应验证

## 微信/企业微信（WeChat/WeCom）数据获取

### 官方API情况

**企业微信（WeCom）**: https://work.weixin.qq.com/

企业微信提供了**完整的官方API**：

1. **企业微信API**: 
   - 完全免费（企业认证后）
   - 提供完整的消息接收能力
   - 支持多种消息类型

2. **普通微信**: 
   - 官方API限制较多
   - 主要用于公众号
   - 个人微信无官方API

### 实现方式

#### 方式1: 企业微信Webhook（推荐）

企业微信会主动推送消息到你的服务器：

```go
// Webhook接收端点
POST /api/webhook/wechat

// 请求参数（GET验证）
GET /api/webhook/wechat?signature=xxx&timestamp=xxx&nonce=xxx&echostr=xxx

// 消息格式（XML或JSON）
<xml>
  <ToUserName>接收者</ToUserName>
  <FromUserName>发送者</FromUserName>
  <CreateTime>时间戳</CreateTime>
  <MsgType>消息类型</MsgType>
  <Content>消息内容</Content>
  <MsgId>消息ID</MsgId>
</xml>
```

**签名验证流程**:
```go
1. 将token、timestamp、nonce按字典序排序
2. 拼接成字符串
3. SHA1加密
4. 与请求参数中的signature对比
5. 验证通过返回echostr
```

#### 方式2: 企业微信API主动拉取（备用）

```go
// 获取access_token
GET https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=xxx&corpsecret=xxx

// 获取消息（需要配置消息接收模式）
// 企业微信支持两种模式：
// 1. 回调模式（推荐）- 主动推送
// 2. API模式 - 主动拉取
```

### 配置步骤

1. **注册企业微信**
   - 访问 https://work.weixin.qq.com/
   - 完成企业认证
   - 创建自建应用

2. **获取凭证**
   - Corp ID（企业ID）
   - Corp Secret（应用密钥）
   - Agent ID（应用ID）

3. **配置接收消息服务器**
   - 在应用管理后台设置
   - URL: `https://your-domain.com/api/webhook/wechat`
   - Token: 自定义令牌
   - EncodingAESKey: 加密密钥（可选）

4. **验证服务器**
   - 企业微信会发送GET请求验证
   - 需要正确响应echostr

## 代码实现详解

### 1. Webhook接收器

```go
// internal/web/routes.go
func handleRedbookWebhook(svc *service.Service) ghttp.HandlerFunc {
    return func(r *ghttp.Request) {
        // 1. 获取签名信息
        signature := r.Header.Get("X-Signature")
        timestamp := r.Header.Get("X-Timestamp")
        nonce := r.Header.Get("X-Nonce")
        
        // 2. 获取请求体
        bodyBytes := r.GetBody()
        
        // 3. 验证签名
        if !svc.VerifyRedbookWebhook(signature, timestamp, nonce, bodyBytes) {
            r.Response.WriteStatus(401, "Invalid signature")
            return
        }
        
        // 4. 处理消息
        svc.ProcessRedbookWebhook(r.GetCtx(), bodyBytes)
    }
}
```

### 2. 消息处理服务

```go
// internal/service/service.go
func (s *Service) ProcessRedbookMessage(ctx context.Context, msg *redbook.Message) error {
    // 1. 查找或创建对话
    conv := &model.Conversation{
        Platform:      "redbook",
        AccountID:     msg.ToUserID,
        CustomerID:    msg.FromUserID,
        Status:        "active",
        StartedAt:     msg.CreatedAt,
        LastMessageAt: msg.CreatedAt,
    }
    s.repo.CreateOrUpdateConversation(ctx, conv)
    
    // 2. 保存消息
    message := &model.Message{
        ConversationID: conv.ID,
        Platform:       "redbook",
        MessageID:      msg.MsgID,
        Content:        msg.Content,
        Timestamp:      msg.CreatedAt,
    }
    s.repo.CreateMessage(ctx, message)
    
    return nil
}
```

### 3. 签名验证

```go
// internal/integration/redbook/client.go
func (c *Client) VerifyWebhook(signature string, timestamp string, nonce string, body []byte) bool {
    // 1. 排序参数
    params := []string{c.config.WebhookToken, timestamp, nonce}
    sort.Strings(params)
    
    // 2. 拼接
    combined := strings.Join(params, "")
    
    // 3. SHA256加密
    hash := sha256.Sum256([]byte(combined))
    expectedSignature := hex.EncodeToString(hash[:])
    
    // 4. 对比
    return signature == expectedSignature
}
```

## 数据存储

### 数据库设计

```sql
-- 对话表
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY,
    platform VARCHAR(20),      -- redbook/wechat
    account_id VARCHAR(100),   -- 账户ID
    customer_id VARCHAR(100),  -- 客户ID
    status VARCHAR(20),         -- active/closed
    started_at TIMESTAMP,
    last_message_at TIMESTAMP,
    message_count INTEGER,
    synced_to_sheet BOOLEAN
);

-- 消息表
CREATE TABLE messages (
    id INTEGER PRIMARY KEY,
    conversation_id INTEGER,
    platform VARCHAR(20),
    message_id VARCHAR(100) UNIQUE,
    sender_id VARCHAR(100),
    sender_type VARCHAR(20),   -- customer/staff
    content TEXT,
    content_type VARCHAR(50),  -- text/image/video
    timestamp TIMESTAMP,
    synced_to_sheet BOOLEAN
);
```

## Google Sheets同步

### 同步机制

```go
// internal/integration/google_sheets/client.go
func (c *Client) StartSyncTask(ctx context.Context, repo *repository.Repository) {
    ticker := time.NewTicker(60 * time.Second)
    
    for {
        select {
        case <-ticker.C:
            // 1. 获取未同步消息
            messages := repo.GetUnsyncedMessages(ctx, 100)
            
            // 2. 同步到Google Sheets
            c.SyncMessages(ctx, messages)
            
            // 3. 标记为已同步
            repo.MarkMessagesAsSynced(ctx, messageIDs)
        }
    }
}
```

## 安全机制

### 1. Webhook签名验证
- 防止伪造请求
- 确保消息来源合法

### 2. HTTPS要求
- 所有Webhook必须使用HTTPS
- 保护数据传输安全

### 3. Token管理
- 定期轮换Token
- 安全存储密钥

## 注意事项

### 小红书API限制

1. **审核要求**: 
   - 需要企业认证
   - 应用需要审核通过

2. **费用**: 
   - 可能有API调用费用
   - 需要查看最新定价

3. **限制**: 
   - 有调用频率限制
   - 需要遵守平台规则

### 企业微信优势

1. **完全免费**: 
   - 企业认证后免费使用
   - 无API调用费用

2. **功能完整**: 
   - 支持所有消息类型
   - 提供完整API

3. **稳定可靠**: 
   - 官方支持
   - 文档完善

## 替代方案

如果官方API不可用，可以考虑：

1. **第三方服务**: 
   - 使用第三方消息转发服务
   - 但需要评估安全性和合规性

2. **爬虫方案**: 
   - 不推荐，违反平台规则
   - 可能被封号

3. **手动导出**: 
   - 定期手动导出数据
   - 适合数据量小的场景

## 总结

- **小红书**: 有官方API，但需要企业认证和审核
- **企业微信**: 有完整的官方API，免费且稳定
- **实现方式**: 主要通过Webhook推送，实时接收消息
- **数据流程**: Webhook → 验证 → 解析 → 存储 → 同步

建议优先使用企业微信，因为其API更完善、免费且稳定。
