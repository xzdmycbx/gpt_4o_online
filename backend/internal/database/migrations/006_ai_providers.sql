-- Migration 006: AI Providers
-- Adds ai_providers table and links ai_models to providers

CREATE TABLE IF NOT EXISTS ai_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    provider_type VARCHAR(50) NOT NULL,  -- 'openai' | 'anthropic' | 'custom'
    api_endpoint TEXT NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    description TEXT NOT NULL DEFAULT '',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_providers_name ON ai_providers(name);
CREATE INDEX IF NOT EXISTS idx_ai_providers_active ON ai_providers(is_active);

-- Add provider_id to ai_models, make api_endpoint/api_key_encrypted optional
ALTER TABLE ai_models
    ADD COLUMN IF NOT EXISTS provider_id UUID REFERENCES ai_providers(id) ON DELETE SET NULL;

-- Make api_endpoint and api_key_encrypted nullable (they become optional when provider_id is set)
ALTER TABLE ai_models
    ALTER COLUMN api_endpoint DROP NOT NULL,
    ALTER COLUMN api_key_encrypted DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ai_models_provider ON ai_models(provider_id);
