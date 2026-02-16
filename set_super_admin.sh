#!/bin/bash

# =============================================================================
# 设置超级管理员脚本
# 用途: 将现有用户设置为超级管理员
# 使用方法: ./set_super_admin.sh <username>
# =============================================================================

# 检查参数
if [ -z "$1" ]; then
    echo "❌ 错误: 请提供用户名"
    echo ""
    echo "使用方法:"
    echo "  docker compose exec app /app/set_super_admin.sh <username>"
    echo ""
    echo "示例:"
    echo "  docker compose exec app /app/set_super_admin.sh alice"
    exit 1
fi

USERNAME="$1"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  设置超级管理员"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "目标用户: $USERNAME"
echo ""

# 确认
read -p "⚠️  确定要将用户 '$USERNAME' 设置为超级管理员吗？[y/N] " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ 操作已取消"
    exit 0
fi

echo ""
echo "正在设置超级管理员..."
echo ""

# 连接数据库并更新用户角色
# 使用参数化查询防止SQL注入
PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}" -v username="$USERNAME" << 'EOF'
UPDATE users
SET role = 'super_admin',
    updated_at = NOW()
WHERE username = :'username'
RETURNING id, username, role, created_at;
EOF

# 检查执行结果
if [ $? -eq 0 ]; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✅ 成功！用户 '$USERNAME' 已设置为超级管理员"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "注意事项:"
    echo "  1. 超级管理员拥有所有权限"
    echo "  2. 可以管理其他管理员"
    echo "  3. 只能通过此脚本设置，后台界面无法设置"
    echo "  4. 请妥善保管超级管理员账号"
    echo ""
else
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "❌ 失败！无法设置超级管理员"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "可能的原因:"
    echo "  1. 用户 '$USERNAME' 不存在"
    echo "  2. 数据库连接失败"
    echo "  3. 权限不足"
    echo ""
    echo "请检查错误信息并重试"
    echo ""
    exit 1
fi
