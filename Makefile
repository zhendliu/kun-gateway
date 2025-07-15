# K8s流量网关构建脚本

# 变量定义
BINARY_NAME=kun-gateway
DOCKER_IMAGE=kun-gateway
VERSION?=latest
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# 构建目标
.PHONY: all build clean test docker-build docker-push deploy

# 默认目标
all: build

# 构建Go二进制文件
build:
	@echo "构建 $(BINARY_NAME)..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/dataplane cmd/dataplane/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/controlplane cmd/controlplane/main.go
	@echo "构建完成"

# 清理构建产物
clean:
	@echo "清理构建产物..."
	rm -rf bin/
	@echo "清理完成"

# 运行测试
test:
	@echo "运行测试..."
	go test ./...
	@echo "测试完成"

# 构建Docker镜像
docker-build:
	@echo "构建Docker镜像..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@echo "Docker镜像构建完成"

# 推送Docker镜像
docker-push:
	@echo "推送Docker镜像..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	@echo "Docker镜像推送完成"

# 部署到K8s
deploy:
	@echo "部署到K8s..."
	kubectl apply -f deployments/
	@echo "部署完成"

# 删除K8s部署
undeploy:
	@echo "删除K8s部署..."
	kubectl delete -f deployments/
	@echo "删除完成"

# 运行数据面（本地开发）
run-dataplane:
	@echo "启动数据面..."
	./bin/dataplane --port=80 --api-port=8080 --log-level=debug

# 运行控制面（本地开发）
run-controlplane:
	@echo "启动控制面..."
	./bin/controlplane --port=9090 --dataplane-url=http://localhost:8080 --log-level=debug

# 运行前端（本地开发）
run-frontend:
	@echo "启动前端..."
	cd web/frontend && npm run dev

# 安装前端依赖
install-frontend:
	@echo "安装前端依赖..."
	cd web/frontend && npm install

# 构建前端
build-frontend:
	@echo "构建前端..."
	cd web/frontend && npm run build

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	@echo "代码格式化完成"

# 代码检查
lint:
	@echo "代码检查..."
	golangci-lint run
	@echo "代码检查完成"

# 生成依赖
deps:
	@echo "更新依赖..."
	go mod tidy
	go mod download
	@echo "依赖更新完成"

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  build          - 构建Go二进制文件"
	@echo "  clean          - 清理构建产物"
	@echo "  test           - 运行测试"
	@echo "  docker-build   - 构建Docker镜像"
	@echo "  docker-push    - 推送Docker镜像"
	@echo "  deploy         - 部署到K8s"
	@echo "  undeploy       - 删除K8s部署"
	@echo "  run-dataplane  - 运行数据面（本地）"
	@echo "  run-controlplane - 运行控制面（本地）"
	@echo "  run-frontend   - 运行前端（本地）"
	@echo "  install-frontend - 安装前端依赖"
	@echo "  build-frontend - 构建前端"
	@echo "  fmt            - 格式化代码"
	@echo "  lint           - 代码检查"
	@echo "  deps           - 更新依赖"
	@echo "  help           - 显示帮助信息" 