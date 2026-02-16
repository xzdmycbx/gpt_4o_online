-- AI Chat System Database Schema
-- PostgreSQL 16

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create ENUM types
CREATE TYPE user_role AS ENUM ('super_admin', 'admin', 'user');
CREATE TYPE permission_type AS ENUM (
    'manage_users',
    'manage_admins',
    'manage_models',
    'manage_settings',
    'view_audit_logs',
    'view_statistics',
    'view_conversations',
    'view_memories'
);
CREATE TYPE memory_category AS ENUM ('preference', 'fact', 'context');
CREATE TYPE email_provider AS ENUM ('smtp', 'resend');
CREATE TYPE audit_action AS ENUM (
    'user_created',
    'user_updated',
    'user_banned',
    'user_unbanned',
    'model_created',
    'model_updated',
    'model_deleted',
    'settings_updated',
    'permission_changed',
    'password_reset'
);

-- ============================================================================
-- USERS TABLE
-- ============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255), -- nullable for OAuth-only users
    display_name VARCHAR(100),
    avatar_url TEXT,
    role user_role NOT NULL DEFAULT 'user',

    -- OAuth2 fields
    oauth2_provider VARCHAR(50), -- 'twitter', 'google', etc.
    oauth2_id VARCHAR(255),
    oauth2_access_token TEXT,
    oauth2_refresh_token TEXT,
    oauth2_token_expiry TIMESTAMP,

    -- Status fields
    is_banned BOOLEAN NOT NULL DEFAULT false,
    ban_reason TEXT,
    banned_at TIMESTAMP,
    banned_by UUID REFERENCES users(id),

    -- Rate limiting override
    rate_limit_exempt BOOLEAN NOT NULL DEFAULT false,
    custom_rate_limit INTEGER, -- requests per minute, NULL means use default

    -- Timestamps
    email_verified_at TIMESTAMP,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT oauth2_unique UNIQUE (oauth2_provider, oauth2_id)
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_oauth2 ON users(oauth2_provider, oauth2_id);
CREATE INDEX idx_users_role ON users(role);

-- ============================================================================
-- PERMISSIONS TABLE
-- ============================================================================
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name permission_type UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default permissions
INSERT INTO permissions (name, description) VALUES
    ('manage_users', 'Create, update, ban/unban regular users'),
    ('manage_admins', 'Manage admin accounts (super_admin only)'),
    ('manage_models', 'Configure AI models and API keys'),
    ('manage_settings', 'Update system settings and configurations'),
    ('view_audit_logs', 'View system audit logs'),
    ('view_statistics', 'View system statistics and token leaderboard'),
    ('view_conversations', 'View user conversations (super_admin only)'),
    ('view_memories', 'View user memories (super_admin only)');

-- ============================================================================
-- ROLE PERMISSIONS TABLE
-- ============================================================================
CREATE TABLE role_permissions (
    role user_role NOT NULL,
    permission permission_type NOT NULL,
    PRIMARY KEY (role, permission),
    FOREIGN KEY (permission) REFERENCES permissions(name) ON DELETE CASCADE
);

-- Assign permissions to roles
-- Super Admin: All permissions
INSERT INTO role_permissions (role, permission) VALUES
    ('super_admin', 'manage_users'),
    ('super_admin', 'manage_admins'),
    ('super_admin', 'manage_models'),
    ('super_admin', 'manage_settings'),
    ('super_admin', 'view_audit_logs'),
    ('super_admin', 'view_statistics'),
    ('super_admin', 'view_conversations'),
    ('super_admin', 'view_memories');

-- Admin: Limited permissions (NO view_conversations, NO view_memories, NO manage_admins)
INSERT INTO role_permissions (role, permission) VALUES
    ('admin', 'manage_users'),
    ('admin', 'manage_models'),
    ('admin', 'manage_settings'),
    ('admin', 'view_audit_logs'),
    ('admin', 'view_statistics');

-- ============================================================================
-- AI MODELS TABLE
-- ============================================================================
CREATE TABLE ai_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL, -- 'openai', 'anthropic', 'custom', etc.
    api_endpoint TEXT NOT NULL,
    api_key_encrypted TEXT NOT NULL, -- AES-256-GCM encrypted
    model_identifier VARCHAR(100) NOT NULL, -- e.g., 'gpt-4', 'claude-3-opus'

    -- Model capabilities
    supports_streaming BOOLEAN NOT NULL DEFAULT true,
    supports_functions BOOLEAN NOT NULL DEFAULT false,
    max_tokens INTEGER NOT NULL DEFAULT 4096,

    -- Pricing (per 1k tokens)
    input_price_per_1k DECIMAL(10, 6),
    output_price_per_1k DECIMAL(10, 6),

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,

    -- Metadata
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_models_provider ON ai_models(provider);
CREATE INDEX idx_ai_models_active ON ai_models(is_active);

-- ============================================================================
-- CONVERSATIONS TABLE
-- ============================================================================
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL DEFAULT 'New Conversation',
    model_id UUID REFERENCES ai_models(id),

    -- Metadata
    message_count INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    last_message_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_user ON conversations(user_id);
CREATE INDEX idx_conversations_updated ON conversations(updated_at DESC);
CREATE INDEX idx_conversations_user_updated ON conversations(user_id, updated_at DESC);

-- ============================================================================
-- MESSAGES TABLE
-- ============================================================================
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL, -- 'user', 'assistant', 'system'
    content TEXT NOT NULL,

    -- Token usage
    input_tokens INTEGER,
    output_tokens INTEGER,
    total_tokens INTEGER,

    -- Model used
    model_id UUID REFERENCES ai_models(id),

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_conversation_created ON messages(conversation_id, created_at);

-- ============================================================================
-- MEMORIES TABLE
-- ============================================================================
CREATE TABLE memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    category memory_category NOT NULL DEFAULT 'context',
    importance INTEGER NOT NULL DEFAULT 5 CHECK (importance >= 1 AND importance <= 10),

    -- Context
    source_conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    source_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,

    -- Usage tracking
    times_used INTEGER NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_memories_user ON memories(user_id);
CREATE INDEX idx_memories_user_importance ON memories(user_id, importance DESC);
CREATE INDEX idx_memories_user_last_used ON memories(user_id, last_used_at DESC);
CREATE INDEX idx_memories_category ON memories(category);

-- ============================================================================
-- TOKEN USAGE TABLE (for leaderboard)
-- ============================================================================
CREATE TABLE token_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model_id UUID REFERENCES ai_models(id) ON DELETE SET NULL,

    -- Token counts
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,

    -- Cost tracking
    estimated_cost DECIMAL(10, 6),

    -- Time period (for monthly/weekly stats)
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, model_id, period_start)
);

CREATE INDEX idx_token_usage_user ON token_usage(user_id);
CREATE INDEX idx_token_usage_period ON token_usage(period_start, period_end);
CREATE INDEX idx_token_usage_total ON token_usage(total_tokens DESC);

-- ============================================================================
-- EMAIL CONFIGURATION TABLE
-- ============================================================================
CREATE TABLE email_config (
    id SERIAL PRIMARY KEY,
    provider email_provider NOT NULL DEFAULT 'smtp',

    -- SMTP settings
    smtp_host VARCHAR(255),
    smtp_port INTEGER,
    smtp_user VARCHAR(255),
    smtp_password_encrypted TEXT, -- AES-256-GCM encrypted
    smtp_use_tls BOOLEAN DEFAULT true,

    -- Resend settings
    resend_api_key_encrypted TEXT, -- AES-256-GCM encrypted

    -- Email settings
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(100),

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_tested_at TIMESTAMP,
    test_result TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- AUDIT LOGS TABLE
-- ============================================================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    action audit_action NOT NULL,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    target_user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Details
    details JSONB,
    ip_address INET,
    user_agent TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_target ON audit_logs(target_user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

-- ============================================================================
-- USER SETTINGS TABLE (Multi-device sync)
-- ============================================================================
CREATE TABLE user_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- UI Settings
    theme VARCHAR(20) DEFAULT 'dark', -- 'dark', 'light', 'auto'
    font_size VARCHAR(20) DEFAULT 'medium', -- 'small', 'medium', 'large'
    language VARCHAR(10) DEFAULT 'en',

    -- Notification Settings
    notifications_enabled BOOLEAN DEFAULT true,
    notification_sound BOOLEAN DEFAULT true,

    -- Chat Preferences
    default_model_id UUID REFERENCES ai_models(id),
    stream_response BOOLEAN DEFAULT true,
    show_token_count BOOLEAN DEFAULT false,

    -- Advanced Settings (JSONB for flexibility)
    advanced_settings JSONB DEFAULT '{}'::jsonb,

    -- Sync metadata
    device_id VARCHAR(255),
    last_synced_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE (user_id)
);

CREATE INDEX idx_user_settings_user ON user_settings(user_id);
CREATE INDEX idx_user_settings_updated ON user_settings(updated_at DESC);

-- ============================================================================
-- TRIGGERS FOR UPDATED_AT
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_memories_updated_at BEFORE UPDATE ON memories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ai_models_updated_at BEFORE UPDATE ON ai_models
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_email_config_updated_at BEFORE UPDATE ON email_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_settings_updated_at BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- TRIGGER FOR CONVERSATION STATISTICS
-- ============================================================================
CREATE OR REPLACE FUNCTION update_conversation_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations
    SET
        message_count = message_count + 1,
        total_tokens = total_tokens + COALESCE(NEW.total_tokens, 0),
        last_message_at = NEW.created_at
    WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_conversation_stats_trigger AFTER INSERT ON messages
    FOR EACH ROW EXECUTE FUNCTION update_conversation_stats();

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- User statistics view
CREATE VIEW user_statistics AS
SELECT
    u.id,
    u.username,
    u.display_name,
    u.avatar_url,
    COUNT(DISTINCT c.id) as conversation_count,
    COUNT(DISTINCT m.id) as message_count,
    COALESCE(SUM(tu.total_tokens), 0) as total_tokens_used,
    COALESCE(SUM(tu.estimated_cost), 0) as total_cost,
    u.created_at as member_since
FROM users u
LEFT JOIN conversations c ON u.id = c.user_id
LEFT JOIN messages m ON c.id = m.conversation_id
LEFT JOIN token_usage tu ON u.id = tu.user_id
WHERE u.role = 'user' AND u.is_banned = false
GROUP BY u.id, u.username, u.display_name, u.avatar_url, u.created_at;

-- Token leaderboard view
CREATE VIEW token_leaderboard AS
SELECT
    u.id,
    u.username,
    u.display_name,
    u.avatar_url,
    SUM(tu.total_tokens) as total_tokens,
    COUNT(*) as total_requests,
    SUM(tu.input_tokens) as input_tokens,
    SUM(tu.output_tokens) as output_tokens,
    SUM(tu.estimated_cost) as estimated_cost,
    RANK() OVER (ORDER BY SUM(tu.total_tokens) DESC) as rank
FROM users u
JOIN token_usage tu ON u.id = tu.user_id
WHERE u.role = 'user' AND u.is_banned = false
GROUP BY u.id, u.username, u.display_name, u.avatar_url
ORDER BY total_tokens DESC;

-- ============================================================================
-- INITIAL DATA
-- ============================================================================

-- Create default email config (needs to be configured by admin)
INSERT INTO email_config (provider, from_email, from_name, is_active)
VALUES ('smtp', 'noreply@change-this-domain.com', 'AI Chat System', false);

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON TABLE users IS 'User accounts with OAuth2 support and role-based access control';
COMMENT ON TABLE permissions IS 'Available system permissions';
COMMENT ON TABLE role_permissions IS 'Mapping of roles to permissions';
COMMENT ON TABLE ai_models IS 'AI model configurations with encrypted API keys';
COMMENT ON TABLE conversations IS 'Chat conversation sessions';
COMMENT ON TABLE messages IS 'Individual messages within conversations';
COMMENT ON TABLE memories IS 'User memory storage extracted from conversations';
COMMENT ON TABLE token_usage IS 'Token usage tracking for statistics and leaderboards';
COMMENT ON TABLE email_config IS 'Email service configuration';
COMMENT ON TABLE audit_logs IS 'Audit trail of system actions';
COMMENT ON TABLE user_settings IS 'User preferences and settings with multi-device sync';
