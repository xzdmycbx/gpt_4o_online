# 多阶段构建：前端 + 后端一体化

# Stage 1: 构建前端
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

# 复制前端文件
COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Stage 2: 构建后端
FROM golang:1.21-alpine AS backend-builder

# 安装构建依赖
RUN apk add --no-cache git make

WORKDIR /build

# 复制 Go 模块文件
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端代码
COPY backend/ ./

# 从前端构建阶段复制构建产物到正确位置（go:embed 需要）
COPY --from=frontend-builder /frontend/dist ./cmd/server/web/dist

# 构建后端主程序（前端已嵌入）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o ai-chat cmd/server/main.go

# 构建超级管理员设置工具
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o set-super-admin cmd/set-super-admin/main.go

# Stage 3: 运行时镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /build/ai-chat .
COPY --from=backend-builder /build/set-super-admin .

# 创建数据目录
RUN mkdir -p /app/data

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./ai-chat"]
