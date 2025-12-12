# 部署指南

## 系统要求

- Go 1.21+
- 数据库：SQLite（开发）/ MySQL或PostgreSQL（生产）
- Google Cloud账号（用于Google Sheets API）
- 公网服务器（用于接收Webhook）

## 部署步骤

### 1. 服务器准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 安装MySQL（可选）
sudo apt install mysql-server -y
```

### 2. 克隆和构建

```bash
# 克隆代码
git clone <repository-url>
cd monitoring-system

# 安装依赖
go mod download

# 构建
go build -o monitoring-system main.go
```

### 3. 配置系统

```bash
# 创建配置目录
mkdir -p config data logs

# 复制配置文件
cp config.yaml config.local.yaml

# 编辑配置
nano config.local.yaml
```

### 4. 设置Google Cloud服务账号

1. 访问 [Google Cloud Console](https://console.cloud.google.com/)
2. 创建新项目或选择现有项目
3. 启用 Google Sheets API
4. 创建服务账号：
   - 转到 "IAM & Admin" > "Service Accounts"
   - 点击 "Create Service Account"
   - 填写名称和描述
   - 授予 "Editor" 角色（或自定义角色）
5. 创建密钥：
   - 点击服务账号
   - 转到 "Keys" 标签
   - 点击 "Add Key" > "Create new key"
   - 选择 JSON 格式
   - 下载并保存到 `config/google-credentials.json`

### 5. 配置Google Sheets

1. 创建新的Google Sheets电子表格
2. 从URL中获取Spreadsheet ID（`https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit`）
3. 分享电子表格给服务账号邮箱（授予"编辑者"权限）
4. 在配置文件中填入Spreadsheet ID

### 6. 配置Redbook Webhook

1. 登录 [小红书开放平台](https://open.xiaohongshu.com/)
2. 进入应用管理
3. 配置Webhook URL: `https://your-domain.com/api/webhook/redbook`
4. 设置Webhook Token
5. 在配置文件中填入App ID、App Secret和Token

### 7. 配置WeChat/WeCom Webhook

#### 企业微信

1. 登录 [企业微信管理后台](https://work.weixin.qq.com/)
2. 进入应用管理 > 自建应用
3. 配置接收消息服务器URL: `https://your-domain.com/api/webhook/wechat`
4. 设置Token和EncodingAESKey
5. 在配置文件中填入Corp ID、Corp Secret、Agent ID和Token

### 8. 使用systemd运行服务

创建服务文件 `/etc/systemd/system/monitoring-system.service`:

```ini
[Unit]
Description=Redbook WeChat Monitoring System
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/monitoring-system
ExecStart=/opt/monitoring-system/monitoring-system -config /opt/monitoring-system/config.local.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable monitoring-system
sudo systemctl start monitoring-system
sudo systemctl status monitoring-system
```

### 9. 配置Nginx反向代理（可选）

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 10. 配置SSL证书（使用Let's Encrypt）

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d your-domain.com
```

## 监控和维护

### 查看日志

```bash
# systemd日志
sudo journalctl -u monitoring-system -f

# 应用日志
tail -f logs/monitoring.log
```

### 备份数据库

```bash
# SQLite
cp data/conversations.db backups/conversations-$(date +%Y%m%d).db

# MySQL
mysqldump -u user -p conversation_logs > backups/conversations-$(date +%Y%m%d).sql
```

### 更新系统

```bash
# 停止服务
sudo systemctl stop monitoring-system

# 备份当前版本
cp monitoring-system monitoring-system.backup

# 拉取最新代码
git pull

# 重新构建
go build -o monitoring-system main.go

# 启动服务
sudo systemctl start monitoring-system
```

## 故障排查

### Webhook不接收消息

1. 检查服务器防火墙是否开放端口
2. 验证Webhook URL可访问性
3. 检查签名验证逻辑
4. 查看应用日志

### Google Sheets同步失败

1. 验证服务账号凭证文件路径
2. 检查服务账号权限
3. 确认Spreadsheet ID正确
4. 查看API配额限制

### 数据库连接失败

1. 检查数据库服务是否运行
2. 验证连接配置
3. 检查数据库用户权限
4. 查看数据库日志

## 安全建议

1. **使用HTTPS**: 所有Webhook必须使用HTTPS
2. **限制访问**: 配置防火墙规则，只允许必要的IP访问
3. **定期更新**: 保持系统和依赖库更新
4. **密钥管理**: 使用密钥管理服务（如AWS Secrets Manager）
5. **监控告警**: 设置异常监控和告警
6. **数据加密**: 敏感数据加密存储
7. **访问审计**: 记录所有管理操作

## 性能优化

1. **数据库索引**: 确保关键字段有索引
2. **连接池**: 配置适当的数据库连接池大小
3. **批量处理**: Google Sheets同步使用批量写入
4. **缓存**: 对频繁查询的数据使用缓存
5. **异步处理**: 使用消息队列处理高并发场景
