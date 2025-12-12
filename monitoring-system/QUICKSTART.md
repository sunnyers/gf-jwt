# 快速开始指南

## 本地运行步骤

### 1. 前置要求

确保已安装：
- **Go 1.21+** - [下载地址](https://go.dev/dl/)
- **Git** - 用于克隆代码

### 2. 获取代码

```bash
# 如果代码在本地，直接进入目录
cd monitoring-system

# 或者从Git仓库克隆
git clone <repository-url>
cd monitoring-system
```

### 3. 配置系统

```bash
# 复制配置文件模板
cp config.example.yaml config.yaml

# 编辑配置文件（使用你喜欢的编辑器）
nano config.yaml
# 或
vim config.yaml
# 或
code config.yaml
```

**最小配置**（用于测试）：
- 保持 `redbook.enabled: false`
- 保持 `wechat.enabled: false`
- 保持 `google_sheets.enabled: false`
- 数据库使用默认SQLite配置

### 4. 运行系统

#### 方式1: 使用启动脚本（推荐）

```bash
# 赋予执行权限
chmod +x run.sh

# 运行
./run.sh
```

#### 方式2: 手动运行

```bash
# 下载依赖
go mod download

# 运行
go run main.go -config config.yaml
```

#### 方式3: 先构建再运行

```bash
# 构建
go build -o monitoring-system main.go

# 运行
./monitoring-system -config config.yaml
```

### 5. 验证运行

打开浏览器访问：
- 健康检查: http://localhost:8080/api/health
- 应该返回: `{"status":"ok"}`

### 6. 测试API

```bash
# 健康检查
curl http://localhost:8080/api/health

# 获取对话列表（需要认证）
curl http://localhost:8080/api/admin/conversations \
  -H "Authorization: Bearer your-token"
```

## 开发模式

### 启用热重载（需要安装air）

```bash
# 安装air
go install github.com/cosmtrek/air@latest

# 运行
air
```

### 查看日志

```bash
# 实时查看日志
tail -f logs/monitoring.log

# 或查看systemd日志（如果使用systemd）
journalctl -u monitoring-system -f
```

## 常见问题

### Q: 端口8080已被占用

**解决**: 修改 `config.yaml` 中的 `server.port` 为其他端口，如 `8081`

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

# 重新下载
go mod download
```

### Q: 编译错误

**解决**: 
```bash
# 清理并重新下载依赖
go clean -modcache
go mod download
go mod tidy
```

## 下一步

1. **配置Redbook集成**（如需要）
   - 在小红书开放平台注册应用
   - 获取App ID和Secret
   - 配置Webhook URL

2. **配置WeChat集成**（如需要）
   - 在企业微信管理后台创建应用
   - 获取企业ID和应用密钥
   - 配置接收消息服务器

3. **配置Google Sheets**（如需要）
   - 创建Google Cloud项目
   - 启用Google Sheets API
   - 创建服务账号并下载凭证
   - 分享电子表格给服务账号

4. **部署到生产环境**
   - 参考 `DEPLOYMENT.md`
   - 配置HTTPS
   - 设置防火墙规则
   - 配置监控和告警

## 获取帮助

- 查看完整文档: `README_zh.md`
- 部署指南: `DEPLOYMENT.md`
- 实施指南: `IMPLEMENTATION_GUIDE.md`
