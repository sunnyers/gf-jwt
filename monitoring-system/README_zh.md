# Redbook和WeChat对话监控系统

这是一个专为加拿大金融科技公司设计的集中式对话监控系统，用于自动记录和管理来自Redbook（小红书）和WeChat（微信/企业微信）的所有客户对话。

## 功能特性

- ✅ **Redbook集成**: 自动接收和记录Redbook商业账户的客户消息
- ✅ **WeChat/WeCom集成**: 支持企业微信和普通微信的消息捕获
- ✅ **自动同步**: 实时将对话记录同步到Google Sheets
- ✅ **Web管理界面**: 提供友好的Web界面查看和管理所有对话
- ✅ **数据持久化**: 使用SQLite/MySQL/PostgreSQL存储所有对话数据
- ✅ **安全认证**: 支持JWT认证保护管理界面
- ✅ **实时监控**: 实时记录所有入站消息和聊天活动

## 系统架构

```
┌─────────────┐      ┌─────────────┐
│   Redbook   │─────▶│  Webhook    │
│   Platform  │      │  接收器      │
└─────────────┘      └─────────────┘
                            │
┌─────────────┐      ┌─────▼─────┐      ┌──────────────┐
│   WeChat    │─────▶│  消息处理  │─────▶│   数据库      │
│   Platform  │      │  服务      │      │  (SQLite/    │
└─────────────┘      └─────┬─────┘      │  MySQL/PG)   │
                            │            └──────┬───────┘
                            │                   │
                     ┌──────▼──────────┐       │
                     │  Google Sheets   │◀──────┘
                     │   自动同步        │
                     └──────────────────┘
```

## 快速开始

### 1. 环境要求

- Go 1.21 或更高版本
- SQLite（默认）或 MySQL/PostgreSQL
- Google Cloud服务账号（用于Google Sheets集成）

### 2. 安装依赖

```bash
cd monitoring-system
go mod download
```

### 3. 配置系统

复制配置文件模板并修改：

```bash
cp config.yaml config.local.yaml
```

编辑 `config.local.yaml`，填入以下信息：

- **Redbook配置**: App ID、App Secret、Webhook Token
- **WeChat配置**: 企业ID、应用密钥、Agent ID
- **Google Sheets配置**: Spreadsheet ID、服务账号凭证文件路径
- **数据库配置**: 根据你的需求选择SQLite、MySQL或PostgreSQL

### 4. 设置Google Sheets

1. 在Google Cloud Console创建项目
2. 启用Google Sheets API
3. 创建服务账号并下载JSON凭证文件
4. 将凭证文件保存到 `config/google-credentials.json`
5. 在Google Sheets中分享电子表格给服务账号邮箱（只读权限）

### 5. 运行系统

```bash
go run main.go -config config.local.yaml
```

系统将在 `http://localhost:8080` 启动。

## 配置说明

### Redbook配置

1. 登录小红书开放平台 (https://open.xiaohongshu.com/)
2. 创建应用并获取App ID和App Secret
3. 配置Webhook URL: `https://your-domain.com/api/webhook/redbook`
4. 设置Webhook Token用于验证

### WeChat/WeCom配置

#### 企业微信（推荐）

1. 登录企业微信管理后台 (https://work.weixin.qq.com/)
2. 创建应用并获取：
   - Corp ID（企业ID）
   - Corp Secret（应用密钥）
   - Agent ID（应用ID）
3. 配置接收消息服务器URL: `https://your-domain.com/api/webhook/wechat`
4. 设置Token用于验证

#### 普通微信

如果使用普通微信，需要配置App ID和App Secret。

## API文档

### Webhook端点

#### Redbook Webhook
```
POST /api/webhook/redbook
Headers:
  X-Signature: <签名>
  X-Timestamp: <时间戳>
  X-Nonce: <随机数>
Body: <消息JSON>
```

#### WeChat Webhook
```
POST /api/webhook/wechat
GET /api/webhook/wechat?signature=...&timestamp=...&nonce=...&echostr=...
```

### 管理API（需要认证）

#### 获取对话列表
```
GET /api/admin/conversations?platform=redbook&account_id=xxx&status=active
Authorization: Bearer <JWT_TOKEN>
```

#### 获取消息列表
```
GET /api/admin/conversations/:id/messages
Authorization: Bearer <JWT_TOKEN>
```

#### 获取统计信息
```
GET /api/admin/stats
Authorization: Bearer <JWT_TOKEN>
```

#### 手动触发同步
```
POST /api/admin/sync
Authorization: Bearer <JWT_TOKEN>
```

## 数据模型

### Conversation（对话）
- ID: 对话ID
- Platform: 平台（redbook/wechat）
- AccountID: 账户ID
- CustomerID: 客户ID
- Status: 状态（active/closed）
- StartedAt: 开始时间
- LastMessageAt: 最后消息时间
- MessageCount: 消息数量

### Message（消息）
- ID: 消息ID
- ConversationID: 对话ID
- Platform: 平台
- MessageID: 平台消息ID
- SenderID: 发送者ID
- SenderType: 发送者类型（customer/staff）
- Content: 消息内容
- ContentType: 内容类型（text/image/video等）
- Timestamp: 时间戳

## 部署建议

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o monitoring-system main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/monitoring-system .
COPY --from=builder /app/config.yaml .
CMD ["./monitoring-system"]
```

### 生产环境注意事项

1. **安全性**:
   - 使用HTTPS
   - 配置强JWT密钥
   - 定期轮换API密钥
   - 限制Webhook IP白名单

2. **性能**:
   - 使用MySQL或PostgreSQL替代SQLite
   - 配置数据库连接池
   - 启用消息队列处理高并发

3. **监控**:
   - 配置日志收集
   - 设置告警规则
   - 监控API调用频率

4. **备份**:
   - 定期备份数据库
   - 备份配置文件
   - 保留Google Sheets历史版本

## 常见问题

### Q: 如何验证Webhook是否正常工作？
A: 检查 `/api/health` 端点，查看日志文件，或在管理界面查看最近的消息。

### Q: Google Sheets同步失败怎么办？
A: 检查服务账号权限、凭证文件路径、Spreadsheet ID是否正确，查看日志获取详细错误信息。

### Q: 如何查看历史对话？
A: 使用管理API或Web界面，支持按平台、账户、时间范围筛选。

### Q: 系统支持哪些消息类型？
A: 目前支持文本、图片、视频等常见类型，具体取决于平台API支持。

## 合规性说明

本系统设计用于合规的对话记录和监控：

1. **数据隐私**: 所有数据存储在您控制的数据库中，Google Sheets仅作为查看界面
2. **访问控制**: 通过JWT认证限制管理界面访问
3. **审计日志**: 记录所有同步操作和错误
4. **数据保留**: 可根据需要配置数据保留策略

## 许可证

MIT License

## 支持

如有问题或建议，请联系技术支持团队。
