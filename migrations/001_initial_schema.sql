-- APM (Assistant Product Manager) Initial Schema
-- Migration: 001_initial_schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- USERS & ORGANIZATIONS
-- ============================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    cellphone VARCHAR(50),

    -- Atlassian OAuth tokens (encrypted with AES)
    atlassian_access_token_encrypted TEXT,
    atlassian_refresh_token_encrypted TEXT,
    atlassian_token_expires_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,

    -- Atlassian configuration
    jira_base_url VARCHAR(500),
    jira_cloud_id VARCHAR(100),
    confluence_base_url VARCHAR(500),
    confluence_cloud_id VARCHAR(100),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE organization_members (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member', -- owner, admin, member
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);

CREATE INDEX idx_org_members_user ON organization_members(user_id);

-- ============================================
-- PRD TEMPLATES
-- ============================================

CREATE TABLE prd_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,

    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- feature, bug_fix, enhancement, technical

    -- Template structure as JSON schema
    content_schema JSONB NOT NULL,

    -- Example content for reference
    example_content JSONB,

    -- System templates are available to all orgs
    is_system_template BOOLEAN DEFAULT FALSE,

    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_templates_org ON prd_templates(organization_id);
CREATE INDEX idx_templates_type ON prd_templates(type);
CREATE INDEX idx_templates_system ON prd_templates(is_system_template) WHERE is_system_template = TRUE;

-- ============================================
-- PRDs (Product Requirements Documents)
-- ============================================

CREATE TABLE prds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    template_id UUID REFERENCES prd_templates(id) ON DELETE SET NULL,

    title VARCHAR(500) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- draft, review, approved, published, archived

    -- Structured PRD content stored as JSONB
    content JSONB NOT NULL DEFAULT '{}',

    -- AI generation metadata
    ai_generated BOOLEAN DEFAULT FALSE,
    source_notes TEXT, -- Original input used for AI generation

    -- Atlassian integration
    jira_epic_key VARCHAR(50),
    jira_epic_id VARCHAR(100),
    confluence_page_id VARCHAR(100),
    confluence_page_url VARCHAR(500),

    -- Ownership
    owner_id UUID NOT NULL REFERENCES users(id),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_prds_org ON prds(organization_id);
CREATE INDEX idx_prds_status ON prds(status);
CREATE INDEX idx_prds_owner ON prds(owner_id);
CREATE INDEX idx_prds_jira_epic ON prds(jira_epic_key) WHERE jira_epic_key IS NOT NULL;
CREATE INDEX idx_prds_created ON prds(created_at DESC);

-- ============================================
-- PRD VERSIONS (History Tracking)
-- ============================================

CREATE TABLE prd_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    prd_id UUID NOT NULL REFERENCES prds(id) ON DELETE CASCADE,

    version_number INTEGER NOT NULL,
    content JSONB NOT NULL,

    changed_by UUID REFERENCES users(id),
    change_summary TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_prd_versions_prd ON prd_versions(prd_id);
CREATE UNIQUE INDEX idx_prd_versions_unique ON prd_versions(prd_id, version_number);

-- ============================================
-- PRIORITY SCORES
-- ============================================

CREATE TABLE priority_scores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    prd_id UUID NOT NULL REFERENCES prds(id) ON DELETE CASCADE,

    framework VARCHAR(50) NOT NULL, -- rice, moscow, ice, weighted

    -- RICE framework components
    reach INTEGER,           -- Users reached per quarter
    impact INTEGER,          -- 0.25=minimal, 0.5=low, 1=medium, 2=high, 3=massive (stored as 25, 50, 100, 200, 300)
    confidence INTEGER,      -- Percentage: 50, 80, 100
    effort INTEGER,          -- Person-weeks

    -- MoSCoW category
    moscow_category VARCHAR(20), -- must, should, could, wont

    -- ICE framework components
    ice_impact INTEGER,      -- 1-10 scale
    ice_confidence INTEGER,  -- 1-10 scale
    ice_ease INTEGER,        -- 1-10 scale

    -- Weighted scoring (custom criteria)
    weighted_criteria JSONB, -- [{name, weight, score}]

    -- Calculated final score
    final_score DECIMAL(10, 2) NOT NULL,

    scored_by UUID REFERENCES users(id),
    scored_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    notes TEXT
);

CREATE INDEX idx_priority_scores_prd ON priority_scores(prd_id);
CREATE INDEX idx_priority_scores_framework ON priority_scores(framework);
CREATE INDEX idx_priority_scores_score ON priority_scores(final_score DESC);
CREATE UNIQUE INDEX idx_priority_scores_unique ON priority_scores(prd_id, framework);

-- ============================================
-- PRODUCT DISCOVERY
-- ============================================

CREATE TABLE discovery_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    type VARCHAR(50) NOT NULL, -- interview, research, opportunity, solution, jtbd
    title VARCHAR(500) NOT NULL,

    -- Structured content
    content JSONB NOT NULL DEFAULT '{}',

    -- Tags for categorization
    tags TEXT[],

    -- Linked PRDs
    linked_prd_ids UUID[],

    -- AI-generated insights
    ai_insights JSONB,

    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_discovery_org ON discovery_items(organization_id);
CREATE INDEX idx_discovery_type ON discovery_items(type);
CREATE INDEX idx_discovery_tags ON discovery_items USING GIN(tags);
CREATE INDEX idx_discovery_linked_prds ON discovery_items USING GIN(linked_prd_ids);

-- ============================================
-- JIRA SYNC LOG
-- ============================================

CREATE TABLE jira_sync_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    prd_id UUID REFERENCES prds(id) ON DELETE SET NULL,

    jira_issue_key VARCHAR(50) NOT NULL,
    jira_issue_id VARCHAR(100),
    issue_type VARCHAR(50), -- epic, story, task, bug

    sync_direction VARCHAR(20) NOT NULL, -- to_jira, from_jira
    sync_status VARCHAR(50) NOT NULL, -- pending, success, failed

    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,

    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_jira_sync_prd ON jira_sync_log(prd_id);
CREATE INDEX idx_jira_sync_issue ON jira_sync_log(jira_issue_key);
CREATE INDEX idx_jira_sync_status ON jira_sync_log(sync_status);

-- ============================================
-- CONFLUENCE SYNC LOG
-- ============================================

CREATE TABLE confluence_sync_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    prd_id UUID REFERENCES prds(id) ON DELETE SET NULL,

    confluence_page_id VARCHAR(100) NOT NULL,
    page_title VARCHAR(500),
    page_version INTEGER,
    space_key VARCHAR(50),

    sync_direction VARCHAR(20) NOT NULL, -- to_confluence, from_confluence
    sync_status VARCHAR(50) NOT NULL, -- pending, success, failed

    error_message TEXT,

    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_confluence_sync_prd ON confluence_sync_log(prd_id);
CREATE INDEX idx_confluence_sync_page ON confluence_sync_log(confluence_page_id);

-- ============================================
-- AI GENERATION LOG
-- ============================================

CREATE TABLE ai_generation_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,

    generation_type VARCHAR(50) NOT NULL, -- prd, story, criteria, summary, refinement

    -- Input/Output
    input_text TEXT,
    prompt_template VARCHAR(100),
    output_text TEXT,

    -- Model information
    provider VARCHAR(50) NOT NULL, -- openai, anthropic
    model_used VARCHAR(100) NOT NULL,

    -- Usage metrics
    input_tokens INTEGER,
    output_tokens INTEGER,
    total_tokens INTEGER,
    latency_ms INTEGER,

    -- Cost tracking (in cents)
    cost_cents INTEGER,

    -- User feedback
    feedback_rating INTEGER, -- 1-5 rating
    feedback_text TEXT,

    -- Result tracking
    was_accepted BOOLEAN, -- Did user keep the generated content?

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ai_log_user ON ai_generation_log(user_id);
CREATE INDEX idx_ai_log_org ON ai_generation_log(organization_id);
CREATE INDEX idx_ai_log_type ON ai_generation_log(generation_type);
CREATE INDEX idx_ai_log_created ON ai_generation_log(created_at DESC);

-- ============================================
-- USER STORIES (Extracted from PRDs)
-- ============================================

CREATE TABLE user_stories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    prd_id UUID NOT NULL REFERENCES prds(id) ON DELETE CASCADE,

    -- Story format
    as_a VARCHAR(255) NOT NULL,        -- "As a [persona]"
    i_want VARCHAR(500) NOT NULL,      -- "I want [capability]"
    so_that VARCHAR(500),              -- "So that [benefit]"

    -- Details
    description TEXT,
    acceptance_criteria JSONB DEFAULT '[]', -- Array of criteria

    -- Estimation
    story_points INTEGER,
    priority VARCHAR(20), -- high, medium, low

    -- Jira sync
    jira_issue_key VARCHAR(50),
    jira_issue_id VARCHAR(100),

    -- Status
    status VARCHAR(50) DEFAULT 'draft', -- draft, ready, in_progress, done

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_stories_prd ON user_stories(prd_id);
CREATE INDEX idx_stories_jira ON user_stories(jira_issue_key) WHERE jira_issue_key IS NOT NULL;

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_prds_updated_at
    BEFORE UPDATE ON prds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_templates_updated_at
    BEFORE UPDATE ON prd_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_discovery_updated_at
    BEFORE UPDATE ON discovery_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_stories_updated_at
    BEFORE UPDATE ON user_stories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- SEED DATA: System Templates
-- ============================================

INSERT INTO prd_templates (id, name, description, type, content_schema, is_system_template) VALUES
(
    uuid_generate_v4(),
    'Feature PRD',
    'Standard template for new feature development',
    'feature',
    '{
        "sections": [
            {"name": "overview", "label": "Overview", "type": "text", "required": true},
            {"name": "problem_statement", "label": "Problem Statement", "type": "text", "required": true},
            {"name": "goals", "label": "Goals", "type": "list", "required": true},
            {"name": "user_personas", "label": "User Personas", "type": "personas", "required": true},
            {"name": "user_stories", "label": "User Stories", "type": "stories", "required": true},
            {"name": "requirements", "label": "Requirements", "type": "requirements", "required": true},
            {"name": "acceptance_criteria", "label": "Acceptance Criteria", "type": "list", "required": true},
            {"name": "out_of_scope", "label": "Out of Scope", "type": "list", "required": false},
            {"name": "assumptions", "label": "Assumptions", "type": "list", "required": false},
            {"name": "dependencies", "label": "Dependencies", "type": "list", "required": false},
            {"name": "success_metrics", "label": "Success Metrics", "type": "metrics", "required": true}
        ]
    }'::jsonb,
    TRUE
),
(
    uuid_generate_v4(),
    'Bug Fix PRD',
    'Template for documenting bug fixes',
    'bug_fix',
    '{
        "sections": [
            {"name": "bug_description", "label": "Bug Description", "type": "text", "required": true},
            {"name": "reproduction_steps", "label": "Steps to Reproduce", "type": "list", "required": true},
            {"name": "expected_behavior", "label": "Expected Behavior", "type": "text", "required": true},
            {"name": "actual_behavior", "label": "Actual Behavior", "type": "text", "required": true},
            {"name": "root_cause", "label": "Root Cause Analysis", "type": "text", "required": false},
            {"name": "proposed_fix", "label": "Proposed Fix", "type": "text", "required": true},
            {"name": "testing_plan", "label": "Testing Plan", "type": "list", "required": true},
            {"name": "rollback_plan", "label": "Rollback Plan", "type": "text", "required": false}
        ]
    }'::jsonb,
    TRUE
),
(
    uuid_generate_v4(),
    'Enhancement PRD',
    'Template for improving existing features',
    'enhancement',
    '{
        "sections": [
            {"name": "current_state", "label": "Current State", "type": "text", "required": true},
            {"name": "proposed_enhancement", "label": "Proposed Enhancement", "type": "text", "required": true},
            {"name": "justification", "label": "Business Justification", "type": "text", "required": true},
            {"name": "user_impact", "label": "User Impact", "type": "text", "required": true},
            {"name": "requirements", "label": "Requirements", "type": "requirements", "required": true},
            {"name": "acceptance_criteria", "label": "Acceptance Criteria", "type": "list", "required": true},
            {"name": "success_metrics", "label": "Success Metrics", "type": "metrics", "required": true}
        ]
    }'::jsonb,
    TRUE
),
(
    uuid_generate_v4(),
    'Technical Spec',
    'Template for technical specifications',
    'technical',
    '{
        "sections": [
            {"name": "overview", "label": "Technical Overview", "type": "text", "required": true},
            {"name": "architecture", "label": "Architecture", "type": "text", "required": true},
            {"name": "api_design", "label": "API Design", "type": "text", "required": false},
            {"name": "data_model", "label": "Data Model", "type": "text", "required": false},
            {"name": "security_considerations", "label": "Security Considerations", "type": "list", "required": true},
            {"name": "performance_requirements", "label": "Performance Requirements", "type": "list", "required": false},
            {"name": "testing_strategy", "label": "Testing Strategy", "type": "text", "required": true},
            {"name": "deployment_plan", "label": "Deployment Plan", "type": "text", "required": true},
            {"name": "monitoring", "label": "Monitoring & Alerts", "type": "list", "required": false}
        ]
    }'::jsonb,
    TRUE
);
