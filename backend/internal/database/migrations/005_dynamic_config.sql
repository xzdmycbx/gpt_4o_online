-- 添加动态配置项到 system_settings 表
-- 支持从管理后台配置 OAuth2、邮件、AI 等系统设置

-- 插入 OAuth2 配置项
INSERT INTO system_settings (setting_key, setting_value, description, value_type) VALUES
    ('oauth2_twitter_enabled', 'false', '是否启用 Twitter OAuth2 登录', 'bool'),
    ('oauth2_twitter_client_id', '', 'Twitter OAuth2 Client ID', 'string'),
    ('oauth2_twitter_client_secret', '', 'Twitter OAuth2 Client Secret（加密存储）', 'string'),
    ('oauth2_twitter_redirect_url', '', 'Twitter OAuth2 Redirect URL', 'string')
ON CONFLICT (setting_key) DO NOTHING;

-- 插入邮件配置项
INSERT INTO system_settings (setting_key, setting_value, description, value_type) VALUES
    ('email_enabled', 'false', '是否启用邮件服务', 'bool'),
    ('email_provider', 'smtp', '邮件提供商（smtp/resend）', 'string'),
    ('email_smtp_host', '', 'SMTP 服务器地址', 'string'),
    ('email_smtp_port', '587', 'SMTP 服务器端口', 'int'),
    ('email_smtp_user', '', 'SMTP 用户名', 'string'),
    ('email_smtp_password', '', 'SMTP 密码（加密存储）', 'string'),
    ('email_from', 'noreply@example.com', '发件人邮箱地址', 'string'),
    ('email_from_name', 'AI Chat System', '发件人显示名称', 'string'),
    ('email_resend_api_key', '', 'Resend API Key（加密存储）', 'string')
ON CONFLICT (setting_key) DO NOTHING;

-- 插入 AI 配置项
INSERT INTO system_settings (setting_key, setting_value, description, value_type) VALUES
    ('ai_default_memory_model', 'gpt-3.5-turbo', '默认记忆提取模型', 'string'),
    ('ai_memory_extraction_enabled', 'true', '是否启用记忆提取功能', 'bool')
ON CONFLICT (setting_key) DO NOTHING;

-- 更新列注释以支持加密类型的说明
COMMENT ON COLUMN system_settings.value_type IS '配置值类型：string(字符串), int(整数), bool(布尔), json(JSON对象)。敏感信息应使用加密存储。';
COMMENT ON COLUMN system_settings.setting_value IS '配置值。敏感信息（如密码、密钥）应加密后存储。';
