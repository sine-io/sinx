.PHONY: build run test clean docker docker-up docker-down

# 构建项目
build:
	go build -o sinx.exe main.go

# 运行项目
run:
	go run main.go

# 运行测试
test:
	go test -v ./...

# 清理构建文件
clean:
	rm -f sinx sinx.exe

# 安装依赖
deps:
	go mod tidy
	go mod download

# 构建Docker镜像
docker:
	docker build -t sinx-app .

# 启动Docker Compose
docker-up:
	docker-compose up -d

# 停止Docker Compose
docker-down:
	docker-compose down

# 查看Docker日志
docker-logs:
	docker-compose logs -f sinx-app

# 代码格式化
fmt:
	go fmt ./...

# 代码检查
vet:
	go vet ./...

# 完整检查（格式化+检查+构建）
check: fmt vet build

# 开发环境启动（本地数据库）
dev: deps check run

# 生产环境部署
deploy: clean build docker docker-up
