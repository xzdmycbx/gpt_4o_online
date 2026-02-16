# AI 聊天系统

一个功能完整的 AI 聊天平台，支持多模型对话、智能记忆系统和精细化权限管理。

## 功能特性

- 🤖 **多模型支持**：兼容 OpenAI 格式的多种 AI 模型
- 🧠 **智能记忆**：自动提取和管理对话记忆
- 🔐 **高级认证**：JWT + OAuth2 (Twitter) 登录
- 👥 **角色权限**：超级管理员、管理员、用户三级权限
- 🌍 **地理过滤**：基于 IP 的访问控制（可阻止中国大陆）
- ⚡ **速率限制**：可配置的默认速率限制 + 单用户自定义限制（Redis 实现）
- 📧 **邮件集成**：支持 SMTP 和 Resend.com
- 📊 **Token 排行榜**：追踪和展示用户的 Token 使用统计
- 📱 **PWA 支持**：可安装、离线缓存、推送通知
- 🔄 **多设备同步**：设置在多设备间无缝同步
- 🎨 **现代 UI**：Telegram 风格响应式设计

## 技术栈

### 后端
- **框架**：Golang 1.21 + Gin
- **数据库**：PostgreSQL 16
- **缓存**：Redis 7
- **认证**：JWT + OAuth2

### 前端
- **框架**：React 18 + TypeScript
- **构建工具**：Vite
- **样式**：Styled Components
- **状态管理**：Zustand + React Query
- **PWA**：Vite PWA Plugin + Workbox

### 部署
- **容器化**：Docker + Docker Compose
- **反向代理**：Nginx（生产环境）

## 快速开始

### ⚠️ 重要说明

**本项目需要在支持 Docker、Go 和 Node.js 的服务器上运行。**

如果您的设备无法运行 Docker 或 Go 命令，请使用以下方式：
1. 使用云服务器（阿里云、腾讯云等）
2. 使用虚拟机（VirtualBox、VMware）
3. 使用 WSL2（Windows Subsystem for Linux）

### 前置要求

- Docker 和 Docker Compose
- Node.js 20+（用于构建前端）
- Go 1.21+（用于构建后端）
- Make（可选，用于便捷命令）
- Linux 或 macOS 环境（推荐）

### 开发环境

1. **克隆仓库**
   ```bash
   git clone <repository-url>
   cd gpt_4o_online
   ```

2. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件配置各项参数
   ```

3. **下载 GeoIP2 数据库**（可选）
   - 在 https://www.maxmind.com/en/geolite2/signup 注册
   - 下载 GeoLite2-Country.mmdb
   - 放置到 `backend/data/` 目录

4. **启动开发环境**
   ```bash
   make dev
   ```
   或者：
   ```bash
   docker-compose -f docker-compose.dev.yml up --build
   ```

5. **访问应用**
   - 前端：http://localhost:3000
   - 后端 API：http://localhost:8080
   - 默认管理员账号在 `.env` 文件中配置

### 生产部署

1. **配置环境**
   ```bash
   cp .env.example .env
   # 使用生产环境配置编辑 .env
   # 重要：修改所有密钥和密码！
   ```

2. **构建并启动**
   ```bash
   make build-all
   make prod
   ```

3. **设置超级管理员**
   超级管理员通过环境变量配置，在 `.env` 文件中设置：
   ```
   SUPER_ADMIN_USERNAME=admin
   SUPER_ADMIN_PASSWORD=your_secure_password
   SUPER_ADMIN_EMAIL=admin@example.com
   ```
   首次启动时会自动创建。

## 项目结构

```
gpt_4o_online/
├── backend/              # Go 后端
│   ├── cmd/server/       # 应用入口点
│   ├── internal/         # 内部包
│   │   ├── api/          # HTTP 处理器和路由
│   │   ├── service/      # 业务逻辑
│   │   ├── repository/   # 数据访问层
│   │   ├── model/        # 数据模型
│   │   └── pkg/          # 工具（JWT、OAuth2 等）
│   └── scripts/          # 工具脚本
├── frontend/             # React 前端
│   ├── src/              # 源代码
│   │   ├── components/   # React 组件
│   │   ├── api/          # API 客户端
│   │   ├── hooks/        # 自定义 Hooks
│   │   └── styles/       # 样式和主题
│   └── public/           # 静态资源
└── docker-compose.yml    # 生产环境配置文件
```

## API 文档

### 认证
- `POST /api/v1/auth/login` - 用户名/密码登录
- `GET /api/v1/auth/oauth2/twitter` - Twitter OAuth2 流程
- `POST /api/v1/auth/forgot-password` - 请求密码重置
- `POST /api/v1/auth/reset-password` - 使用令牌重置密码

### 聊天
- `GET /api/v1/conversations` - 列出对话
- `POST /api/v1/conversations` - 创建新对话
- `POST /api/v1/conversations/:id/messages` - 发送消息
- `WS /api/v1/chat/stream` - WebSocket 流式响应

### 记忆
- `GET /api/v1/memories` - 列出用户记忆
- `PUT /api/v1/memories/:id` - 更新记忆
- `DELETE /api/v1/memories/:id` - 删除记忆

### 管理
- `GET /api/v1/admin/users` - 列出用户
- `PUT /api/v1/admin/users/:id/ban` - 封禁/解禁用户
- `GET /api/v1/admin/models` - 列出 AI 模型
- `POST /api/v1/admin/models` - 添加 AI 模型
- `GET /api/v1/admin/statistics/tokens` - Token 排行榜

## 配置

### 环境变量

查看 `.env.example` 了解所有可用的配置选项。

关键配置：
- `JWT_SECRET`：JWT 签名密钥（生产环境必须修改！）
- `ENCRYPTION_KEY`：32 字节的 AES-256 加密密钥
- `OAUTH2_TWITTER_CLIENT_ID`：Twitter OAuth2 客户端 ID
- `OAUTH2_TWITTER_CLIENT_SECRET`：Twitter OAuth2 客户端密钥
- `GEOIP_BLOCK_CHINA`：启用/禁用中国 IP 阻止
- `RATE_LIMIT_DEFAULT_PER_MINUTE`：默认速率限制（应用于所有用户，可在后台为每个用户单独配置）
- `SUPER_ADMIN_USERNAME`：超级管理员用户名（首次启动自动创建）
- `SUPER_ADMIN_PASSWORD`：超级管理员密码
- `SUPER_ADMIN_EMAIL`：超级管理员邮箱

### 数据库迁移

数据库迁移在首次启动时自动运行。手动运行：

```bash
make db-migrate
```

重置数据库（警告：删除所有数据）：

```bash
make db-reset
```

## 安全特性

- **JWT 认证**：HMAC-SHA256 签名令牌
- **密码哈希**：bcrypt 自动加盐
- **API 密钥加密**：AES-256-GCM 加密敏感数据
- **CORS 保护**：白名单配置
- **XSS 保护**：Content-Security-Policy 头部
- **CSRF 保护**：OAuth2 state 参数
- **SQL 注入防护**：参数化查询
- **审计日志**：记录所有敏感操作

## 开发

### 后端开发

```bash
cd backend
go run cmd/server/main.go
```

使用热重载（Air）：
```bash
cd backend
air
```

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

### 运行测试

```bash
make test
```

## Makefile 命令

- `make help` - 显示所有可用命令
- `make init` - 初始化项目
- `make dev` - 启动开发环境
- `make build-all` - 构建前端和后端
- `make prod` - 启动生产环境
- `make clean` - 清理构建产物

## PWA 功能

前端是一个渐进式 Web 应用（PWA），具有：
- **安装到主屏幕**：支持移动端和桌面端
- **离线支持**：缓存资源实现离线访问
- **推送通知**：（即将推出）
- **后台同步**：连接恢复时同步数据
- **响应式设计**：移动优先，适配所有设备

## 贡献

欢迎贡献！请遵循以下指南：
1. Fork 仓库
2. 创建功能分支
3. 进行修改
4. 充分测试
5. 提交 Pull Request

## 许可证

本项目采用 MIT 许可证。

## 支持

如有问题、疑问或贡献，请在 GitHub 上提交 Issue。

## 路线图

- [ ] 向量数据库集成（Qdrant）优化记忆
- [ ] 多语言支持（i18n）
- [ ] 图片/文件上传支持
- [ ] 语音输入/输出
- [ ] 高级分析仪表板
- [ ] 社区功能（分享对话）
- [ ] 插件系统

---

为 AI 社区用心打造 ❤️
