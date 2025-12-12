# 本地运行指南

## ✅ 项目已准备就绪

项目已经成功编译，可以直接在本地运行！

## 快速开始（3步）

### 1. 准备配置文件

```bash
cd monitoring-system

# 复制配置文件模板
cp config.example.yaml config.yaml

# 编辑配置文件（最小配置即可运行）
nano config.yaml
```

**最小配置**（用于测试）：
- 保持所有 `enabled: false`（Redbook、WeChat、Google Sheets）
- 数据库使用默认SQLite配置
- 服务器端口：8080

### 2. 运行系统

**方式1: 使用启动脚本（推荐）**
```bash
chmod +x run.sh
./run.sh
```

**方式2: 直接运行**
```bash
go run main.go -config config.yaml
```

**方式3: 先构建再运行**
```bash
go build -o monitoring-system main.go
./monitoring-system -config config.yaml
```

### 3. 验证运行

打开浏览器访问：
```
http://localhost:8080/api/health
```

应该返回：
```json
{"status":"ok"}
```

## 项目结构

```
monitoring-system/
├── main.go                    # 程序入口
├── config.yaml               # 配置文件（需要创建）
├── config.example.yaml       # 配置文件模板
├── run.sh                    # 启动脚本
├── go.mod                    # Go模块定义
├── go.sum                    # 依赖锁定文件
├── internal/                 # 内部包
│   ├── cmd/                  # 命令处理
│   ├── config/               # 配置管理
│   ├── database/             # 数据库
│   ├── model/                # 数据模型
│   ├── repository/           # 数据访问
│   ├── service/              # 业务逻辑
│   ├── integration/          # 第三方集成
│   │   ├── redbook/          # Redbook API
│   │   ├── wechat/           # WeChat API
│   │   └── google_sheets/    # Google Sheets API
│   └── web/                  # Web路由
├── data/                     # 数据库文件（自动创建）
├── logs/                     # 日志文件（自动创建）
└── config/                   # 配置文件目录
```

## 测试API

### 健康检查
```bash
curl http://localhost:8080/api/health
```

### 获取对话列表（需要认证）
```bash
curl http://localhost:8080/api/admin/conversations \
  -H "Authorization: Bearer your-token"
```

## 下一步配置

### 配置Redbook（如需要）

1. 在小红书开放平台注册应用
2. 获取App ID和App Secret
3. 在 `config.yaml` 中设置：
   ```yaml
   redbook:
     enabled: true
     app_id: "your-app-id"
     app_secret: "your-app-secret"
     webhook_token: "your-webhook-token"
     webhook_url: "https://your-domain.com/api/webhook/redbook"
   ```

### 配置WeChat（如需要）

1. 在企业微信管理后台创建应用
2. 获取Corp ID、Corp Secret、Agent ID
3. 在 `config.yaml` 中设置：
   ```yaml
   wechat:
     enabled: true
     wecom:
       corp_id: "your-corp-id"
       corp_secret: "your-corp-secret"
       agent_id: "your-agent-id"
       webhook_token: "your-webhook-token"
       webhook_url: "https://your-domain.com/api/webhook/wechat"
   ```

### 配置Google Sheets（如需要）

1. 创建Google Cloud项目
2. 启用Google Sheets API
3. 创建服务账号并下载JSON凭证
4. 将凭证文件保存到 `config/google-credentials.json`
5. 在 `config.yaml` 中设置：
   ```yaml
   google_sheets:
     enabled: true
     spreadsheet_id: "your-spreadsheet-id"
     credentials_file: "./config/google-credentials.json"
     sheet_name: "对话记录"
     sync_interval: 60
   ```

## 常见问题

### Q: 端口8080被占用
**解决**: 修改 `config.yaml` 中的 `server.port` 为其他端口

### Q: 数据库文件权限错误
**解决**: 
```bash
mkdir -p data
chmod 755 data
```

### Q: 依赖下载失败
**解决**: 
```bash
# 设置Go代理（中国用户）
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

### Q: 编译错误
**解决**: 
```bash
go clean -modcache
go mod download
go mod tidy
go build -o monitoring-system main.go
```

## 开发提示

### 查看日志
```bash
tail -f logs/monitoring.log
```

### 热重载开发（需要安装air）
```bash
go install github.com/cosmtrek/air@latest
air
```

### 数据库查看（SQLite）
```bash
sqlite3 data/conversations.db
.tables
SELECT * FROM conversations;
```

## 文档

- **快速开始**: `QUICKSTART.md`
- **完整文档**: `README_zh.md`
- **部署指南**: `DEPLOYMENT.md`
- **实施指南**: `IMPLEMENTATION_GUIDE.md`

## 支持

如有问题，请查看日志文件或参考文档。
