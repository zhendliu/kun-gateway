# 多阶段构建Dockerfile

# 第一阶段：构建Go应用
FROM golang:1.20-alpine AS go-builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建数据面和控制面
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dataplane cmd/dataplane/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o controlplane cmd/controlplane/main.go

# 第二阶段：构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# 复制前端文件
COPY web/frontend/package*.json ./
RUN npm install

COPY web/frontend/ ./
RUN npm run build

# 第三阶段：最终镜像
FROM alpine:latest

# 安装必要的工具
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 从第一阶段复制二进制文件
COPY --from=go-builder /app/dataplane .
COPY --from=go-builder /app/controlplane .

# 从第二阶段复制前端文件
COPY --from=frontend-builder /app/dist ./web/dist

# 设置权限
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 80 8080 9090

# 默认启动数据面
CMD ["./dataplane"] 