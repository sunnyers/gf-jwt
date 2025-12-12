.PHONY: build run test clean deps install

# 构建应用
build:
	go build -o monitoring-system main.go

# 运行应用
run:
	go run main.go -config config.yaml

# 运行测试
test:
	go test ./...

# 清理构建文件
clean:
	rm -f monitoring-system
	rm -rf data/*.db
	rm -rf logs/*.log

# 下载依赖
deps:
	go mod download
	go mod tidy

# 安装依赖
install: deps
	go install ./...

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 构建Docker镜像
docker-build:
	docker build -t monitoring-system:latest .

# 运行Docker容器
docker-run:
	docker run -d \
		--name monitoring-system \
		-p 8080:8080 \
		-v $(PWD)/config.local.yaml:/app/config.yaml \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/logs:/app/logs \
		monitoring-system:latest

# 停止Docker容器
docker-stop:
	docker stop monitoring-system
	docker rm monitoring-system
