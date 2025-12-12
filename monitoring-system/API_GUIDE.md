# 官方API使用指南

## 一、小红书（Redbook）官方API

### 1. 官方平台

**小红书开放平台**: https://open.xiaohongshu.com/

### 2. API类型

#### A. 消息推送API（Webhook）

**用途**: 接收客户发送给商家的消息

**特点**:
- ✅ 实时推送，无需轮询
- ✅ 官方支持
- ✅ 安全可靠

**申请条件**:
1. 企业认证账号
2. 商业账户
3. 开发者审核通过

**API端点**:
```
POST https://your-domain.com/api/webhook/redbook
```

**请求格式**:
```json
{
  "event_type": "message",
  "timestamp": 1234567890,
  "data": {
    "msg_id": "消息唯一ID",
    "from_user_id": "客户用户ID",
    "to_user_id": "商家账户ID",
    "content": "消息内容",
    "msg_type": "text|image|video",
    "timestamp": 1234567890
  }
}
```

**签名验证**:
```go
// 1. 获取请求头
X-Signature: <签名>
X-Timestamp: <时间戳>
X-Nonce: <随机数>

// 2. 验证算法
token + timestamp + nonce → 排序 → SHA256 → 对比signature
```

#### B. 主动拉取API（备用）

**用途**: 如果Webhook不可用，可以主动拉取消息

**API端点**:
```
GET https://api.xiaohongshu.com/v1/messages
```

**需要参数**:
- `access_token`: 访问令牌
- `limit`: 每页数量
- `cursor`: 游标（分页）

**获取access_token**:
```
POST https://api.xiaohongshu.com/oauth/access_token
{
  "app_id": "你的App ID",
  "app_secret": "你的App Secret",
  "grant_type": "client_credentials"
}
```

### 3. 申请流程

1. **注册开发者账号**
   - 访问 https://open.xiaohongshu.com/
   - 使用企业账号登录
   - 完成开发者认证

2. **创建应用**
   - 填写应用信息
   - 选择应用类型（商业服务）
   - 提交审核

3. **配置Webhook**
   - 设置Webhook URL（必须是HTTPS）
   - 设置Webhook Token（用于签名验证）
   - 保存配置

4. **获取凭证**
   - App ID
   - App Secret
   - Webhook Token

5. **测试连接**
   - 平台会发送测试消息
   - 验证签名是否正确
   - 确认消息接收正常

### 4. 注意事项

⚠️ **重要限制**:
- 需要企业认证（个人账号无法申请）
- 应用需要审核通过
- 可能有API调用费用（需查看最新定价）
- 有调用频率限制
- Webhook URL必须使用HTTPS

📚 **官方文档**:
- API文档: https://open.xiaohongshu.com/document/api
- 开发者指南: https://open.xiaohongshu.com/guide

---

## 二、企业微信（WeCom）官方API

### 1. 官方平台

**企业微信管理后台**: https://work.weixin.qq.com/

### 2. API类型

#### A. 接收消息API（Webhook - 推荐）

**用途**: 接收客户发送给企业微信应用的消息

**特点**:
- ✅ 完全免费（企业认证后）
- ✅ 实时推送
- ✅ 官方支持，稳定可靠
- ✅ 支持所有消息类型

**申请条件**:
1. 企业微信账号
2. 企业认证（免费）
3. 创建自建应用

**API端点**:
```
POST https://your-domain.com/api/webhook/wechat
GET  https://your-domain.com/api/webhook/wechat  (验证用)
```

**验证请求（GET）**:
```
GET /api/webhook/wechat?signature=xxx&timestamp=xxx&nonce=xxx&echostr=xxx
```

**消息格式（XML）**:
```xml
<xml>
  <ToUserName><![CDATA[企业账号]]></ToUserName>
  <FromUserName><![CDATA[客户账号]]></FromUserName>
  <CreateTime>1234567890</CreateTime>
  <MsgType><![CDATA[text]]></MsgType>
  <Content><![CDATA[消息内容]]></Content>
  <MsgId>消息ID</MsgId>
</xml>
```

**签名验证**:
```go
// 1. 获取参数
signature, timestamp, nonce, echostr

// 2. 验证算法
token + timestamp + nonce → 排序 → SHA1 → 对比signature

// 3. 验证通过返回echostr
```

#### B. 主动发送消息API

**用途**: 主动向客户发送消息

**API端点**:
```
POST https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=xxx
```

**消息类型**:
- 文本消息
- 图片消息
- 语音消息
- 视频消息
- 文件消息
- 图文消息

#### C. 获取access_token

**API端点**:
```
GET https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=xxx&corpsecret=xxx
```

**响应**:
```json
{
  "errcode": 0,
  "errmsg": "ok",
  "access_token": "xxx",
  "expires_in": 7200
}
```

### 3. 配置流程

1. **注册企业微信**
   - 访问 https://work.weixin.qq.com/
   - 注册企业账号
   - 完成企业认证（免费）

2. **创建自建应用**
   - 进入"应用管理"
   - 点击"创建应用"
   - 填写应用信息
   - 获取Agent ID

3. **获取企业凭证**
   - Corp ID（企业ID）
   - 在应用详情中获取Corp Secret（应用密钥）

4. **配置接收消息服务器**
   - 进入应用详情
   - 找到"接收消息"设置
   - 设置服务器URL: `https://your-domain.com/api/webhook/wechat`
   - 设置Token（自定义，用于签名验证）
   - 设置EncodingAESKey（可选，用于消息加密）

5. **验证服务器**
   - 点击"保存"后，企业微信会发送GET请求验证
   - 服务器需要正确响应echostr
   - 验证通过后配置生效

6. **测试消息接收**
   - 向企业微信应用发送测试消息
   - 查看服务器日志确认接收

### 4. 优势

✅ **完全免费**: 企业认证后所有API免费使用
✅ **功能完整**: 支持所有消息类型和功能
✅ **稳定可靠**: 官方支持，文档完善
✅ **实时推送**: Webhook机制，消息实时到达
✅ **安全可靠**: 签名验证，消息加密支持

### 5. 官方文档

📚 **API文档**: https://developer.work.weixin.qq.com/document/path/90239
📚 **接收消息指南**: https://developer.work.weixin.qq.com/document/path/90238
📚 **发送消息指南**: https://developer.work.weixin.qq.com/document/path/90236

---

## 三、普通微信（个人微信/公众号）

### 1. 个人微信

❌ **无官方API**: 个人微信不提供官方API接口

⚠️ **注意事项**:
- 使用第三方工具可能违反微信使用协议
- 存在封号风险
- 不推荐使用

### 2. 微信公众号

✅ **有官方API**: 但主要用于公众号运营

**适用场景**:
- 公众号消息接收
- 公众号消息推送
- 不适合个人微信对话

**API文档**: https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html

---

## 四、实现对比

| 特性 | 小红书 | 企业微信 | 个人微信 |
|------|--------|----------|----------|
| 官方API | ✅ 有 | ✅ 有 | ❌ 无 |
| 免费使用 | ⚠️ 可能收费 | ✅ 免费 | N/A |
| 企业认证 | ✅ 需要 | ✅ 需要 | N/A |
| Webhook支持 | ✅ 支持 | ✅ 支持 | N/A |
| 实时推送 | ✅ 支持 | ✅ 支持 | N/A |
| 消息类型 | 文本/图片/视频 | 全部类型 | N/A |
| 稳定性 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | N/A |
| 推荐度 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐ |

---

## 五、推荐方案

### 最佳方案：企业微信

**理由**:
1. ✅ 完全免费
2. ✅ API完善稳定
3. ✅ 官方支持
4. ✅ 功能完整
5. ✅ 文档详细

### 备选方案：小红书

**适用场景**:
- 主要客户在小红书平台
- 已通过企业认证
- 预算允许API费用

---

## 六、代码示例

### 小红书Webhook接收

```go
// 1. 接收Webhook请求
POST /api/webhook/redbook
Headers:
  X-Signature: <签名>
  X-Timestamp: <时间戳>
  X-Nonce: <随机数>

// 2. 验证签名
func VerifyWebhook(signature, timestamp, nonce string, body []byte) bool {
    params := []string{token, timestamp, nonce}
    sort.Strings(params)
    combined := strings.Join(params, "")
    hash := sha256.Sum256([]byte(combined))
    expectedSignature := hex.EncodeToString(hash[:])
    return signature == expectedSignature
}

// 3. 解析消息
func ParseMessage(body []byte) (*Message, error) {
    var payload WebhookPayload
    json.Unmarshal(body, &payload)
    // 解析data字段
    var msg Message
    json.Unmarshal(payload.Data, &msg)
    return &msg, nil
}
```

### 企业微信Webhook接收

```go
// 1. 验证请求（GET）
GET /api/webhook/wechat?signature=xxx&timestamp=xxx&nonce=xxx&echostr=xxx

func VerifyWebhook(signature, timestamp, nonce, echostr string) (string, bool) {
    params := []string{token, timestamp, nonce}
    sort.Strings(params)
    combined := strings.Join(params, "")
    hash := sha1.Sum([]byte(combined))
    expectedSignature := hex.EncodeToString(hash[:])
    if signature == expectedSignature {
        return echostr, true  // 返回echostr
    }
    return "", false
}

// 2. 接收消息（POST）
POST /api/webhook/wechat
Content-Type: application/xml

// 3. 解析XML消息
func ParseMessage(body []byte) (*Message, error) {
    // 解析XML格式
    // 或使用JSON格式（如果配置了）
}
```

---

## 七、常见问题

### Q1: 小红书API收费吗？
A: 需要查看最新定价政策，可能有API调用费用。

### Q2: 企业微信需要付费吗？
A: 企业认证后完全免费使用。

### Q3: 个人微信有API吗？
A: 没有官方API，不推荐使用第三方工具。

### Q4: Webhook必须用HTTPS吗？
A: 是的，所有平台都要求HTTPS。

### Q5: 如何测试Webhook？
A: 平台会发送测试消息，或使用curl手动测试。

---

## 总结

- **小红书**: 有官方API，需要企业认证，可能有费用
- **企业微信**: 有完整官方API，免费且稳定（强烈推荐）
- **个人微信**: 无官方API，不推荐使用

**建议**: 优先使用企业微信，功能完善且免费。
