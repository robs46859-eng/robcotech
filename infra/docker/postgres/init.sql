-- FullStackArkham Database Initialization
-- This script creates the core schema for the platform

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "vector";

-- ============================================================================
-- TENANTS - Multi-tenant organization isolation
-- ============================================================================

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    api_key_hash VARCHAR(255) NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    quota_monthly BIGINT NOT NULL DEFAULT 10000,
    quota_used BIGINT NOT NULL DEFAULT 0,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_api_key_hash ON tenants(api_key_hash);
CREATE INDEX idx_tenants_status ON tenants(status);

-- ============================================================================
-- USERS - User accounts within tenants
-- ============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);

-- ============================================================================
-- API KEYS - API key management for tenants
-- ============================================================================

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    permissions JSONB NOT NULL DEFAULT '["infer", "read"]',
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);

-- ============================================================================
-- INFERENCE LOGS - All inference requests for billing and audit
-- ============================================================================

CREATE TABLE inference_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    request_id UUID NOT NULL,
    model VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,
    total_tokens INTEGER NOT NULL,
    cost_usd DECIMAL(12, 6) NOT NULL,
    latency_ms INTEGER NOT NULL,
    cache_hit BOOLEAN NOT NULL DEFAULT false,
    escalated BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inference_logs_tenant_id ON inference_logs(tenant_id);
CREATE INDEX idx_inference_logs_created_at ON inference_logs(created_at);
CREATE INDEX idx_inference_logs_model ON inference_logs(model);
CREATE INDEX idx_inference_logs_request_id ON inference_logs(request_id);

-- ============================================================================
-- WORKFLOWS - Orchestration workflow instances
-- ============================================================================

CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    current_step VARCHAR(100),
    state JSONB NOT NULL DEFAULT '{}',
    workflow_state JSONB NOT NULL DEFAULT '{}',
    operational_state JSONB NOT NULL DEFAULT '{}',
    cognitive_state JSONB NOT NULL DEFAULT '{}',
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workflows_tenant_id ON workflows(tenant_id);
CREATE INDEX idx_workflows_status ON workflows(status);
CREATE INDEX idx_workflows_created_at ON workflows(created_at);

-- ============================================================================
-- WORKFLOW STEPS - Individual steps within workflows
-- ============================================================================

CREATE TABLE workflow_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    step_name VARCHAR(100) NOT NULL,
    step_order INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    input JSONB,
    output JSONB,
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workflow_steps_workflow_id ON workflow_steps(workflow_id);
CREATE INDEX idx_workflow_steps_status ON workflow_steps(status);

-- ============================================================================
-- SEMANTIC CACHE - Cached inference responses
-- ============================================================================

CREATE TABLE semantic_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    request_hash VARCHAR(64) NOT NULL,
    request_embedding TEXT,  -- Store as TEXT since pgvector may not be available
    response JSONB NOT NULL,
    model VARCHAR(100) NOT NULL,
    task_type VARCHAR(50),
    similarity_threshold FLOAT NOT NULL DEFAULT 0.95,
    hit_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, request_hash)
);

CREATE INDEX idx_semantic_cache_tenant_id ON semantic_cache(tenant_id);
CREATE INDEX idx_semantic_cache_request_hash ON semantic_cache(request_hash);
CREATE INDEX idx_semantic_cache_model ON semantic_cache(model);

-- ============================================================================
-- MEMORY NOTES - A-MEM evolving memory system
-- ============================================================================

CREATE TABLE memory_notes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID REFERENCES users(id),
    workflow_id UUID REFERENCES workflows(id),
    note_type VARCHAR(50) NOT NULL, -- user, task, workflow, domain, operational
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    embedding VECTOR(1536),
    tags TEXT[] NOT NULL DEFAULT '{}',
    links TEXT[] NOT NULL DEFAULT '{}',
    importance FLOAT NOT NULL DEFAULT 0.5,
    decay_rate FLOAT NOT NULL DEFAULT 0.01,
    accessed_count INTEGER NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_memory_notes_tenant_id ON memory_notes(tenant_id);
CREATE INDEX idx_memory_notes_user_id ON memory_notes(user_id);
CREATE INDEX idx_memory_notes_type ON memory_notes(note_type);
CREATE INDEX idx_memory_notes_tags ON memory_notes USING GIN(tags);

-- ============================================================================
-- ARKHAM SECURITY - Security events and attacker fingerprints
-- ============================================================================

CREATE TABLE arkham_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    event_type VARCHAR(50) NOT NULL, -- benign, probe, attack, scanner
    request_id UUID NOT NULL,
    source_ip INET NOT NULL,
    fingerprint_hash VARCHAR(64),
    threat_score FLOAT NOT NULL DEFAULT 0.0,
    deception_engaged BOOLEAN NOT NULL DEFAULT false,
    blocked BOOLEAN NOT NULL DEFAULT false,
    cross_tenant_shared BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_arkham_events_tenant_id ON arkham_events(tenant_id);
CREATE INDEX idx_arkham_events_event_type ON arkham_events(event_type);
CREATE INDEX idx_arkham_events_source_ip ON arkham_events(source_ip);
CREATE INDEX idx_arkham_events_fingerprint ON arkham_events(fingerprint_hash);
CREATE INDEX idx_arkham_events_created_at ON arkham_events(created_at);

-- ============================================================================
-- ATTACKER FINGERPRINTS - Cross-tenant attacker identification
-- ============================================================================

CREATE TABLE attacker_fingerprints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fingerprint_hash VARCHAR(64) UNIQUE NOT NULL,
    behavior_patterns JSONB NOT NULL DEFAULT '[]',
    first_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    attack_count INTEGER NOT NULL DEFAULT 0,
    blocked_tenants UUID[] NOT NULL DEFAULT '{}',
    threat_level VARCHAR(50) NOT NULL DEFAULT 'unknown'
);

CREATE INDEX idx_attacker_fingerprints_hash ON attacker_fingerprints(fingerprint_hash);
CREATE INDEX idx_attacker_fingerprints_threat_level ON attacker_fingerprints(threat_level);

-- ============================================================================
-- BILLING - Usage metering and billing records
-- ============================================================================

CREATE TABLE billing_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    tokens_used BIGINT NOT NULL DEFAULT 0,
    requests_count BIGINT NOT NULL DEFAULT 0,
    cache_hits BIGINT NOT NULL DEFAULT 0,
    total_cost_usd DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    stripe_customer_id VARCHAR(100),
    stripe_invoice_id VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_billing_records_tenant_id ON billing_records(tenant_id);
CREATE INDEX idx_billing_records_period ON billing_records(period_start, period_end);
CREATE INDEX idx_billing_records_status ON billing_records(status);

-- ============================================================================
-- AUDIT LOG - Comprehensive audit trail
-- ============================================================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    old_value JSONB,
    new_value JSONB,
    source_ip INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- ============================================================================
-- INSERT DEFAULT DATA
-- ============================================================================

-- Insert default tenant (for development)
INSERT INTO tenants (id, name, slug, api_key_hash, plan, quota_monthly)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Default Tenant',
    'default',
    '$2a$10$placeholder_hash_for_dev_api_key',
    'free',
    100000
)
ON CONFLICT (id) DO NOTHING;

-- Insert default API key (for development: fsa_dev_key_12345)
-- In production, this should be properly hashed
INSERT INTO api_keys (tenant_id, name, key_hash, key_prefix, permissions)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Development Key',
    'placeholder_hash',
    'fsa_dev',
    '["infer", "read", "write", "admin"]'
)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SUBSCRIPTION PLANS - Subscription tier definitions
-- ============================================================================

CREATE TABLE subscription_plans (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_monthly DECIMAL(10, 2) NOT NULL DEFAULT 0,
    quota_monthly BIGINT NOT NULL DEFAULT 10000,
    features TEXT[] NOT NULL DEFAULT '{}',
    stripe_price_id VARCHAR(100),
    stripe_yearly_price_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add stripe columns to tenants table
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(100);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(100);

-- Add tenant foreign key to billing_records
ALTER TABLE billing_records 
    DROP CONSTRAINT IF EXISTS billing_records_tenant_id_fkey,
    ADD CONSTRAINT billing_records_tenant_id_fkey 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

-- ============================================================================
-- PAPABASE CRM - Leads and Tasks for Family Studio Suite
-- ============================================================================

CREATE TABLE crm_leads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    company VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'lead', -- lead, quote, scheduled, invoiced, done
    source VARCHAR(100),
    notes TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_crm_leads_tenant_id ON crm_leads(tenant_id);
CREATE INDEX idx_crm_leads_status ON crm_leads(status);
CREATE INDEX idx_crm_leads_email ON crm_leads(email);

CREATE TABLE crm_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    lead_id UUID REFERENCES crm_leads(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, in_progress, completed, cancelled
    priority VARCHAR(50) NOT NULL DEFAULT 'medium', -- low, medium, high, urgent
    assignee VARCHAR(255),
    due_date TIMESTAMP WITH TIME ZONE,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_crm_tasks_tenant_id ON crm_tasks(tenant_id);
CREATE INDEX idx_crm_tasks_lead_id ON crm_tasks(lead_id);
CREATE INDEX idx_crm_tasks_status ON crm_tasks(status);
CREATE INDEX idx_crm_tasks_due_date ON crm_tasks(due_date);

