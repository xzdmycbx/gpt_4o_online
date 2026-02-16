#!/bin/bash

# =============================================================================
# 查看所有超级管理员脚本
# 用途: 列出系统中所有的超级管理员
# 使用方法: ./list_super_admins.sh
# =============================================================================

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  系统超级管理员列表"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 连接数据库并查询所有超级管理员
PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}" << 'EOF'
SELECT
    username,
    email,
    created_at,
    updated_at,
    CASE
        WHEN oauth2_provider IS NOT NULL THEN 'OAuth2 (' || oauth2_provider || ')'
        WHEN password_hash IS NOT NULL THEN 'Password'
        ELSE 'Unknown'
    END as auth_method
FROM users
WHERE role = 'super_admin'
ORDER BY created_at ASC;
EOF

# 检查执行结果
if [ $? -eq 0 ]; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "提示:"
    echo "  • 系统允许设置多个超级管理员"
    echo "  • 使用 set_super_admin.sh <username> 添加新的超级管理员"
    echo "  • 超级管理员可以管理所有用户和系统设置"
    echo ""
else
    echo ""
    echo "❌ 查询失败！请检查数据库连接"
    echo ""
    exit 1
fi
