-- 系统设置表
CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL,
    description TEXT,
    value_type VARCHAR(50) DEFAULT 'string', -- string, int, bool, json
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插入默认系统设置
INSERT INTO system_settings (setting_key, setting_value, description, value_type) VALUES
    ('rate_limit_default_per_minute', '20', '默认速率限制（每分钟消息数）', 'int'),
    ('system_name', 'AI Chat System', '系统名称', 'string'),
    ('maintenance_mode', 'false', '维护模式', 'bool')
ON CONFLICT (setting_key) DO NOTHING;

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_system_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_system_settings_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_system_settings_updated_at();

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings(setting_key);
