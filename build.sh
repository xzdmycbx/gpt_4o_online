#!/bin/bash
# 本地构建脚本

set -e

echo "================================================"
echo "AI Chat 完整构建脚本"
echo "================================================"

# 1. 构建前端
echo ""
echo "步骤 1/3: 构建前端..."
cd frontend
npm install
npm run build
cd ..

# 2. 复制前端产物到后端
echo ""
echo "步骤 2/3: 复制前端产物..."
mkdir -p backend/cmd/server/web/dist
cp -r frontend/dist/* backend/cmd/server/web/dist/
echo "前端产物已复制到: backend/cmd/server/web/dist/"

# 3. 构建后端
echo ""
echo "步骤 3/3: 构建后端..."
cd backend
go build -o ../bin/ai-chat ./cmd/server
cd ..

echo ""
echo "================================================"
echo "✅ 构建完成！"
echo "================================================"
echo "二进制文件位置: bin/ai-chat"
echo ""
echo "运行应用："
echo "  ./bin/ai-chat"
echo ""
echo "或使用 Docker："
echo "  docker-compose build"
echo "  docker-compose up -d"
echo "================================================"
